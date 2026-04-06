package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/awd-platform/awd-arena/internal/config"
	"github.com/awd-platform/awd-arena/internal/middleware"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // allow non-browser clients (e.g. curl, ws cli)
		}
		// Allow same-origin and configured allowed origins
		host := r.Host
		if strings.HasPrefix(origin, "http://"+host) || strings.HasPrefix(origin, "https://"+host) {
			return true
		}
		// Allow common localhost variants for development
		for _, allowed := range []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		} {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// wsConn wraps a websocket connection with a write mutex for concurrent safety.
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// WSHub manages WebSocket connections and game subscriptions.
type WSHub struct {
	mu       sync.RWMutex
	clients  map[*wsConn]struct{}
	gameSub  map[string]map[*wsConn]struct{}
	userSubs map[string][]string // userID -> list of gameIDs they were subscribed to
	userConn map[string]*wsConn  // userID -> current active connection
}

// Hub is the global WebSocket hub instance.
var Hub = &WSHub{
	clients:  make(map[*wsConn]struct{}),
	gameSub:  make(map[string]map[*wsConn]struct{}),
	userSubs: make(map[string][]string),
	userConn: make(map[string]*wsConn),
}

func (h *WSHub) Register(wc *wsConn, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[wc] = struct{}{}
	h.userConn[userID] = wc
}

func (h *WSHub) Unregister(wc *wsConn, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, wc)
	delete(h.userConn, userID)
	for gameID := range h.gameSub {
		delete(h.gameSub[gameID], wc)
	}
}

func (h *WSHub) Subscribe(gameID string, wc *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.gameSub[gameID] == nil {
		h.gameSub[gameID] = make(map[*wsConn]struct{})
	}
	h.gameSub[gameID][wc] = struct{}{}
}

func (h *WSHub) TrackUserSub(userID string, gameID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, gid := range h.userSubs[userID] {
		if gid == gameID {
			return
		}
	}
	h.userSubs[userID] = append(h.userSubs[userID], gameID)
}

func (h *WSHub) RestoreSubs(userID string, wc *wsConn) {
	h.mu.Lock()
	gameIDs := make([]string, len(h.userSubs[userID]))
	copy(gameIDs, h.userSubs[userID])
	h.mu.Unlock()

	for _, gid := range gameIDs {
		h.Subscribe(gid, wc)
	}
}

func (h *WSHub) BroadcastToGame(gameID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for wc := range h.gameSub[gameID] {
		wc.mu.Lock()
		err := wc.conn.WriteMessage(websocket.TextMessage, message)
		wc.mu.Unlock()
		if err != nil {
			logger.Error("ws write error", "error", err)
		}
	}
}

func (h *WSHub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for wc := range h.clients {
		wc.mu.Lock()
		err := wc.conn.WriteMessage(websocket.TextMessage, message)
		wc.mu.Unlock()
		if err != nil {
			logger.Error("ws broadcast error", "error", err)
		}
	}
}

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// HandleWebSocket upgrades HTTP to WebSocket.
func HandleWebSocket(c fiber.Ctx) error {
	return adaptor.HTTPHandlerFunc(wsHandler)(c)
}

// validateWSToken validates JWT from query parameter.
func validateWSToken(tokenString string) (*middleware.Claims, error) {
	secret := config.C.Server.JWTSecret
	token, err := jwt.ParseWithClaims(tokenString, &middleware.Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*middleware.Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Validate JWT from query parameter
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := validateWSToken(tokenString)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	userID := fmt.Sprintf("%d", claims.UserID)
	logger.Info("ws connection authenticated", "user", claims.Username)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("ws upgrade error", "error", err)
		return
	}
	wc := &wsConn{conn: conn}
	defer conn.Close()
	Hub.Register(wc, userID)
	defer Hub.Unregister(wc, userID)
	Hub.RestoreSubs(userID, wc)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			continue
		}
		switch wsMsg.Type {
		case "subscribe":
			if ch, ok := wsMsg.Data.(map[string]interface{}); ok {
				if channel, ok := ch["channel"].(string); ok {
					Hub.Subscribe(channel, wc)
					Hub.TrackUserSub(userID, channel)
				}
			}
		case "unsubscribe":
			// handle unsubscribe
		}
	}
}
