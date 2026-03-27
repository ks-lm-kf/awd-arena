package eventbus

// Handler processes domain events.
type Handler struct {
	bus *Bus
}

// RegisterAll registers all event handlers.
func (h *Handler) RegisterAll() {
	// TODO: register event handlers
}
