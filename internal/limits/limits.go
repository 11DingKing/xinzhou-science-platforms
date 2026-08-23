package limits

import (
	"errors"
	"sync"
)

var ErrExceeded = errors.New("limit exceeded")

type Quota struct {
	mu       sync.Mutex
	capacity int
	used     int
}

func New(capacity int) *Quota {
	if capacity < 0 {
		capacity = 0
	}
	return &Quota{capacity: capacity}
}
func (q *Quota) Reserve(n int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if n < 1 || q.used+n > q.capacity {
		return ErrExceeded
	}
	q.used += n
	return nil
}
func (q *Quota) Release(n int) {
	q.mu.Lock()
	q.used -= n
	if q.used < 0 {
		q.used = 0
	}
	q.mu.Unlock()
}
func (q *Quota) Used() int      { q.mu.Lock(); defer q.mu.Unlock(); return q.used }
func (q *Quota) Available() int { q.mu.Lock(); defer q.mu.Unlock(); return q.capacity - q.used }
func (q *Quota) Resize(capacity int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if capacity < q.used {
		return ErrExceeded
	}
	q.capacity = capacity
	return nil
}
