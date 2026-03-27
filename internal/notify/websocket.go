package notify

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	URL               string        // WebSocket 服务器 URL
	ReconnectInterval time.Duration // 重连间隔
	PingInterval      time.Duration // Ping 间隔
	WriteTimeout      time.Duration // 写超时
}

// WebSocketNotifier WebSocket 通知器
type WebSocketNotifier struct {
	config    WebSocketConfig
	conn      *websocket.Conn
	mu        sync.RWMutex
	connected bool
	stopChan  chan struct{}
}

// NewWebSocketNotifier 创建 WebSocket 通知器
func NewWebSocketNotifier(config WebSocketConfig) *WebSocketNotifier {
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}
	if config.PingInterval == 0 {
		config.PingInterval = 30 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}

	wn := &WebSocketNotifier{
		config:   config,
		stopChan: make(chan struct{}),
	}

	// 启动连接管理
	go wn.manageConnection()

	return wn
}

// Name 返回通知器名称
func (wn *WebSocketNotifier) Name() string {
	return "websocket"
}

// manageConnection 管理连接
func (wn *WebSocketNotifier) manageConnection() {
	for {
		select {
		case <-wn.stopChan:
			return
		default:
			if !wn.isConnected() {
				wn.connect()
			}
			time.Sleep(wn.config.ReconnectInterval)
		}
	}
}

// connect 建立连接
func (wn *WebSocketNotifier) connect() error {
	wn.mu.Lock()
	defer wn.mu.Unlock()

	// 关闭旧连接
	if wn.conn != nil {
		wn.conn.Close()
	}

	// 建立新连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wn.config.URL, nil)
	if err != nil {
		wn.connected = false
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	wn.conn = conn
	wn.connected = true

	// 启动 ping 保持连接
	go wn.pingLoop()

	return nil
}

// isConnected 检查是否已连接
func (wn *WebSocketNotifier) isConnected() bool {
	wn.mu.RLock()
	defer wn.mu.RUnlock()
	return wn.connected
}

// pingLoop 发送 ping 保持连接
func (wn *WebSocketNotifier) pingLoop() {
	ticker := time.NewTicker(wn.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wn.mu.Lock()
			if wn.conn != nil {
				if err := wn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					wn.connected = false
					wn.conn.Close()
					wn.conn = nil
				}
			}
			wn.mu.Unlock()
		case <-wn.stopChan:
			return
		}
	}
}

// Send 发送告警到 WebSocket
func (wn *WebSocketNotifier) Send(alert Alert) error {
	if !wn.isConnected() {
		// 尝试重连
		if err := wn.connect(); err != nil {
			return fmt.Errorf("websocket not connected: %w", err)
		}
	}

	// 序列化告警
	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("marshal alert failed: %w", err)
	}

	wn.mu.Lock()
	defer wn.mu.Unlock()

	if wn.conn == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	// 设置写超时
	if err := wn.conn.SetWriteDeadline(time.Now().Add(wn.config.WriteTimeout)); err != nil {
		wn.connected = false
		return fmt.Errorf("set write deadline failed: %w", err)
	}

	// 发送消息
	if err := wn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		wn.connected = false
		wn.conn.Close()
		wn.conn = nil
		return fmt.Errorf("write message failed: %w", err)
	}

	return nil
}

// Close 关闭连接
func (wn *WebSocketNotifier) Close() error {
	close(wn.stopChan)
	
	wn.mu.Lock()
	defer wn.mu.Unlock()

	if wn.conn != nil {
		return wn.conn.Close()
	}
	
	return nil
}

// SendMessage 发送自定义消息（用于测试）
func (wn *WebSocketNotifier) SendMessage(message interface{}) error {
	if !wn.isConnected() {
		if err := wn.connect(); err != nil {
			return fmt.Errorf("websocket not connected: %w", err)
		}
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	wn.mu.Lock()
	defer wn.mu.Unlock()

	if wn.conn == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	if err := wn.conn.SetWriteDeadline(time.Now().Add(wn.config.WriteTimeout)); err != nil {
		return fmt.Errorf("set write deadline failed: %w", err)
	}

	if err := wn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		wn.connected = false
		wn.conn.Close()
		wn.conn = nil
		return fmt.Errorf("write message failed: %w", err)
	}

	return nil
}
