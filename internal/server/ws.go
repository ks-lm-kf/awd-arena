package server

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHub manages WebSocket connections and game subscriptions.
type WSHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
	gameSub map[string]map[*websocket.Conn]struct{}
}

// Hub is the global WebSocket hub instance.
var Hub = &WSHub{
	clients: make(map[*websocket.Conn]struct{}),
	gameSub: make(map[string]map[*websocket.Conn]struct{}),
}

func (h *WSHub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
}

func (h *WSHub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	for gameID := range h.gameSub {
		delete(h.gameSub[gameID], conn)
	}
}

func (h *WSHub) Subscribe(gameID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.gameSub[gameID] == nil {
		h.gameSub[gameID] = make(map[*websocket.Conn]struct{})
	}
	h.gameSub[gameID][conn] = struct{}{}
}

func (h *WSHub) BroadcastToGame(gameID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.gameSub[gameID] {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			logger.Error("ws write error", "error", err)
		}
	}
}

func (h *WSHub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
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
	logger.Info("ws connection authenticated", "user", claims.Username)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("ws upgrade error", "error", err)
		return
	}
	defer conn.Close()
	Hub.Register(conn)
	defer Hub.Unregister(conn)

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
					Hub.Subscribe(channel, conn)
				}
			}
		case "unsubscribe":
			// handle unsubscribe
		}
	}
}
