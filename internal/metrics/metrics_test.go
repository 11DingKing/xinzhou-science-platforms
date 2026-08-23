package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestCountersAndHistogram(t *testing.T) {
	c := NewCounter()
	c.Inc("complaints")
	c.Add("complaints", 2)
	if c.Get("complaints") != 3 {
		t.Fatal("counter")
	}
	snap := c.Snapshot()
	snap["complaints"] = 0
	if c.Get("complaints") != 3 {
		t.Fatal("snapshot alias")
	}
	h := NewHistogram()
	h.Observe("db", time.Millisecond)
	h.Observe("db", 2*time.Millisecond)
	if !strings.Contains(h.String("db"), "count=2") {
		t.Fatal("histogram")
	}
}
