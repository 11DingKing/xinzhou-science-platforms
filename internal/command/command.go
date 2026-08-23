package command

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Handler func(context.Context, any) error
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func New() *Registry { return &Registry{handlers: map[string]Handler{}} }
func (r *Registry) Register(name string, h Handler) error {
	if name == "" || h == nil {
		return errors.New("invalid command")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.handlers[name]; ok {
		return errors.New("command already registered")
	}
	r.handlers[name] = h
	return nil
}
func (r *Registry) Execute(ctx context.Context, name string, payload any) error {
	r.mu.RLock()
	h, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("command %s not found", name)
	}
	return h(ctx, payload)
}
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		out = append(out, name)
	}
	return out
}

type Envelope struct {
	ID        string
	Name      string
	Payload   any
	CreatedAt time.Time
	Attempts  int
}

func (e Envelope) Valid() bool { return e.ID != "" && e.Name != "" && !e.CreatedAt.IsZero() }
