package worker

import (
	"context"
	"time"
)

type Backoff struct {
	Base, Max time.Duration
	Attempts  int
}

func (b Backoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := b.Base
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= b.Max {
			return b.Max
		}
	}
	if d > b.Max {
		return b.Max
	}
	return d
}
func (b Backoff) Wait(ctx context.Context, attempt int) error {
	timer := time.NewTimer(b.Delay(attempt))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (b Backoff) Exhausted(attempt int) bool { return attempt >= b.Attempts }
