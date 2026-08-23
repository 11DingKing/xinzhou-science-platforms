package queue

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrClosed = errors.New("queue closed")

type Item[T any] struct {
	Value      T
	EnqueuedAt time.Time
}
type Queue[T any] struct {
	mu     sync.Mutex
	items  []Item[T]
	closed bool
	wake   chan struct{}
}

func New[T any]() *Queue[T] { return &Queue[T]{wake: make(chan struct{}, 1)} }
func (q *Queue[T]) Push(v T) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return ErrClosed
	}
	q.items = append(q.items, Item[T]{Value: v, EnqueuedAt: time.Now().UTC()})
	select {
	case q.wake <- struct{}{}:
	default:
	}
	return nil
}
func (q *Queue[T]) Pop(ctx context.Context) (Item[T], error) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			item := q.items[0]
			q.items = q.items[1:]
			q.mu.Unlock()
			return item, nil
		}
		closed := q.closed
		q.mu.Unlock()
		if closed {
			return Item[T]{}, ErrClosed
		}
		select {
		case <-ctx.Done():
			return Item[T]{}, ctx.Err()
		case <-q.wake:
		}
	}
}
func (q *Queue[T]) Close()   { q.mu.Lock(); q.closed = true; close(q.wake); q.mu.Unlock() }
func (q *Queue[T]) Len() int { q.mu.Lock(); defer q.mu.Unlock(); return len(q.items) }
func (q *Queue[T]) Drain() []Item[T] {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := append([]Item[T](nil), q.items...)
	q.items = nil
	return out
}
