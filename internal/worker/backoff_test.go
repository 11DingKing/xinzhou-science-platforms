package worker

import (
	"context"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	b := Backoff{Base: time.Millisecond, Max: 2 * time.Millisecond, Attempts: 3}
	if b.Delay(1) != time.Millisecond || b.Delay(4) != 2*time.Millisecond || !b.Exhausted(3) {
		t.Fatal("backoff")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Wait(ctx, 1); err == nil {
		t.Fatal("cancel")
	}
}
