package eventbus

import "encoding/json"

// Broadcaster is an interface for sending messages without direct server import.
type Broadcaster interface {
	Broadcast(data []byte)
	BroadcastToGame(gameID string, data []byte)
}

// DefaultBroadcaster is a no-op broadcaster (used before server sets real one).
type DefaultBroadcaster struct{}

func (d *DefaultBroadcaster) Broadcast(data []byte)           {}
func (d *DefaultBroadcaster) BroadcastToGame(gameID string, data []byte) {}

// global broadcaster instance
var wsBroadcaster Broadcaster = &DefaultBroadcaster{}

// SetBroadcaster sets the WebSocket broadcaster.
func SetBroadcaster(b Broadcaster) {
	wsBroadcaster = b
}

// BroadcastWS sends data to all WebSocket clients.
func BroadcastWS(data []byte) {
	wsBroadcaster.Broadcast(data)
}

// BroadcastToGameWS sends data to clients subscribed to a game.
func BroadcastToGameWS(gameID string, data []byte) {
	wsBroadcaster.BroadcastToGame(gameID, data)
}

// BroadcastJSON broadcasts a JSON message.
func BroadcastJSON(msgType string, data interface{}) {
	b, _ := json.Marshal(map[string]interface{}{"type": msgType, "data": data})
	BroadcastWS(b)
}
