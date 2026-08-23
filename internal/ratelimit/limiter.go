package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Bucket struct {
	Remaining int
	ResetAt   time.Time
}
type Limiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	items  map[string]Bucket
	now    func() time.Time
}

func New(limit int, window time.Duration) *Limiter {
	if limit < 1 {
		limit = 1
	}
	return &Limiter{limit: limit, window: window, items: map[string]Bucket{}, now: time.Now}
}
func (l *Limiter) Allow(key string) (bool, Bucket) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b := l.items[key]
	if b.ResetAt.IsZero() || !now.Before(b.ResetAt) {
		b = Bucket{Remaining: l.limit, ResetAt: now.Add(l.window)}
	}
	if b.Remaining <= 0 {
		l.items[key] = b
		return false, b
	}
	b.Remaining--
	l.items[key] = b
	return true, b
}
func (l *Limiter) Wait(ctx context.Context, key string) error {
	for {
		ok, b := l.Allow(key)
		if ok {
			return nil
		}
		timer := time.NewTimer(time.Until(b.ResetAt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}
func (l *Limiter) Snapshot(key string) Bucket { l.mu.Lock(); defer l.mu.Unlock(); return l.items[key] }
