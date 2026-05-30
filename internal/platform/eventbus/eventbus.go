package eventbus

import "context"

// Event is a domain event published inside the process.
type Event interface {
	Name() string
}

// Handler processes a single event type.
type Handler func(ctx context.Context, event Event) error

// Bus dispatches events to registered handlers synchronously.
type Bus struct {
	handlers map[string][]Handler
}

// New creates an empty in-process event bus.
func New() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for an event name.
func (b *Bus) Subscribe(eventName string, handler Handler) {
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish invokes all handlers for the event. Business modules will use this after commits.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	for _, h := range b.handlers[event.Name()] {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
