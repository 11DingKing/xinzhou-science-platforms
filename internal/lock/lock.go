package lock

import (
	"context"
	"sync"
	"time"
)

type Keyed struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func New() *Keyed { return &Keyed{locks: map[string]*sync.Mutex{}} }
func (k *Keyed) mutex(key string) *sync.Mutex {
	k.mu.Lock()
	defer k.mu.Unlock()
	m := k.locks[key]
	if m == nil {
		m = &sync.Mutex{}
		k.locks[key] = m
	}
	return m
}
func (k *Keyed) With(ctx context.Context, key string, fn func() error) error {
	m := k.mutex(key)
	done := make(chan struct{})
	go func() { m.Lock(); close(done) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	defer m.Unlock()
	return fn()
}
func (k *Keyed) Try(key string, fn func() error) bool {
	m := k.mutex(key)
	locked := make(chan struct{}, 1)
	go func() { m.Lock(); locked <- struct{}{} }()
	select {
	case <-locked:
		defer m.Unlock()
		_ = fn()
		return true
	case <-time.After(time.Millisecond):
		return false
	}
}
