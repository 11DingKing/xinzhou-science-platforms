package metrics

import (
	"fmt"
	"sync"
	"time"
)

type Counter struct {
	mu     sync.RWMutex
	values map[string]int64
}

func NewCounter() *Counter                  { return &Counter{values: map[string]int64{}} }
func (c *Counter) Add(name string, n int64) { c.mu.Lock(); c.values[name] += n; c.mu.Unlock() }
func (c *Counter) Inc(name string)          { c.Add(name, 1) }
func (c *Counter) Get(name string) int64    { c.mu.RLock(); defer c.mu.RUnlock(); return c.values[name] }
func (c *Counter) Snapshot() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := map[string]int64{}
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

type Timing struct {
	Count int64
	Total time.Duration
	Max   time.Duration
}
type Histogram struct {
	mu     sync.RWMutex
	values map[string]Timing
}

func NewHistogram() *Histogram { return &Histogram{values: map[string]Timing{}} }
func (h *Histogram) Observe(name string, d time.Duration) {
	h.mu.Lock()
	v := h.values[name]
	v.Count++
	v.Total += d
	if d > v.Max {
		v.Max = d
	}
	h.values[name] = v
	h.mu.Unlock()
}
func (h *Histogram) String(name string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	v := h.values[name]
	if v.Count == 0 {
		return ""
	}
	return fmt.Sprintf("count=%d total=%s max=%s", v.Count, v.Total, v.Max)
}
