package eventbus

import (
	"context"
	"sync"
)

type Event struct {
	Topic   string
	Payload any
}
type Handler func(context.Context, Event) error
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func New() *Bus { return &Bus{handlers: map[string][]Handler{}} }
func (b *Bus) Subscribe(topic string, h Handler) {
	b.mu.Lock()
	b.handlers[topic] = append(b.handlers[topic], h)
	b.mu.Unlock()
}
func (b *Bus) Publish(ctx context.Context, e Event) error {
	b.mu.RLock()
	hs := append([]Handler(nil), b.handlers[e.Topic]...)
	b.mu.RUnlock()
	for _, h := range hs {
		if err := h(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
func (b *Bus) Topics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0, len(b.handlers))
	for k := range b.handlers {
		out = append(out, k)
	}
	return out
}
