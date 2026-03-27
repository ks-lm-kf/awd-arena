package eventbus

import "context"

// Bus provides event publishing and subscribing.
type Bus struct{}

// Publish publishes an event.
func (b *Bus) Publish(ctx context.Context, subject string, data interface{}) error {
return nil
}

// Subscribe subscribes to events on a subject.
func (b *Bus) Subscribe(ctx context.Context, subject string, handler func(interface{})) error {
return nil
}

// Close closes the bus connection.
func (b *Bus) Close() error { return nil }

// defaultBus is the global bus instance
var defaultBus = &Bus{}

// GetBus returns the global bus instance
func GetBus() *Bus {
return defaultBus
}
