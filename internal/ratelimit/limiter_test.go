package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestLimiterWindow(t *testing.T) {
	l := New(2, time.Minute)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first denied")
	}
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("second denied")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Fatal("third accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Wait(ctx, "a"); err == nil {
		t.Fatal("cancel ignored")
	}
}
