package eventbus

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/pkg/logger"
)

// Bus provides event publishing and subscribing with real handler dispatch and WebSocket broadcast.
type Bus struct {
	mu            sync.RWMutex
	subscriptions map[string][]func(interface{})
}

// Publish publishes an event to all subscribers of the subject and broadcasts via WebSocket.
func (b *Bus) Publish(ctx context.Context, subject string, data interface{}) error {
	b.mu.RLock()
	handlers := make([]func(interface{}), len(b.subscriptions[subject]))
	copy(handlers, b.subscriptions[subject])
	b.mu.RUnlock()

	// Broadcast via WebSocket
	go func() {
		msg := map[string]interface{}{
			"type": subject,
			"data": data,
			"ts":   time.Now().Unix(),
		}
		payload, err := json.Marshal(msg)
		if err != nil {
			logger.Error("eventbus marshal error", "error", err)
			return
		}
		BroadcastWS(payload)
	}()

	// Call registered handlers
	for _, handler := range handlers {
		func(h func(interface{})) {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("eventbus handler panic", "subject", subject, "panic", r)
					}
				}()
				h(data)
			}()
		}(handler)
	}

	return nil
}

// Subscribe subscribes to events on a subject.
func (b *Bus) Subscribe(ctx context.Context, subject string, handler func(interface{})) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriptions == nil {
		b.subscriptions = make(map[string][]func(interface{}))
	}
	b.subscriptions[subject] = append(b.subscriptions[subject], handler)
	return nil
}

// SubscribeSimple subscribes without a context.
func (b *Bus) SubscribeSimple(subject string, handler func(interface{})) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriptions == nil {
		b.subscriptions = make(map[string][]func(interface{}))
	}
	b.subscriptions[subject] = append(b.subscriptions[subject], handler)
}

// Close closes the bus and clears subscriptions.
func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscriptions = nil
	return nil
}

// defaultBus is the global bus instance
var defaultBus = &Bus{
	subscriptions: make(map[string][]func(interface{})),
}

// GetBus returns the global bus instance
func GetBus() *Bus {
	return defaultBus
}
