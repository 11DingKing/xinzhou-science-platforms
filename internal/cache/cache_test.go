package cache

import (
	"testing"
	"time"
)

func TestCacheExpiry(t *testing.T) {
	now := time.Now()
	c := New[string](time.Minute)
	c.Set("a", "value", now)
	if got, ok := c.Get("a", now.Add(time.Second)); !ok || got != "value" {
		t.Fatal("get")
	}
	if _, ok := c.Get("a", now.Add(time.Minute)); ok {
		t.Fatal("expired")
	}
	if c.Purge(now.Add(time.Minute)) != 1 || c.Len() != 0 {
		t.Fatal("purge")
	}
	c.Set("b", "x", now)
	c.Delete("b")
	if c.Len() != 0 {
		t.Fatal("delete")
	}
}
