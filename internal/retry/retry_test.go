package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySucceedsAndExhausts(t *testing.T) {
	p := Policy{Attempts: 3, Base: time.Millisecond, Max: time.Millisecond}
	calls := 0
	if err := Run(context.Background(), p, func(context.Context) error {
		calls++
		if calls < 2 {
			return errors.New("temporary")
		}
		return nil
	}); err != nil || calls != 2 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	calls = 0
	if err := Run(context.Background(), p, func(context.Context) error { calls++; return errors.New("bad") }); err == nil || calls != 3 {
		t.Fatalf("exhaust err=%v calls=%d", err, calls)
	}
	if p.Delay(10) != time.Millisecond {
		t.Fatal("max delay")
	}
}
