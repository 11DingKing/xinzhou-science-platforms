package clock

import (
	"sync"
	"time"
)

type Clock struct {
	mu  sync.RWMutex
	now time.Time
}

func New(now time.Time) *Clock           { return &Clock{now: now} }
func (c *Clock) Now() time.Time          { c.mu.RLock(); defer c.mu.RUnlock(); return c.now }
func (c *Clock) Set(now time.Time)       { c.mu.Lock(); c.now = now; c.mu.Unlock() }
func (c *Clock) Advance(d time.Duration) { c.mu.Lock(); c.now = c.now.Add(d); c.mu.Unlock() }
func (c *Clock) UTC() time.Time          { return c.Now().UTC() }
