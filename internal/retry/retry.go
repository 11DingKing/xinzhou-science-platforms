package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

var ErrExhausted = errors.New("retry exhausted")

type Policy struct {
	Attempts  int
	Base, Max time.Duration
	Jitter    bool
}

func Default() Policy { return Policy{Attempts: 5, Base: 50 * time.Millisecond, Max: 2 * time.Second} }
func (p Policy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := float64(p.Base) * math.Pow(2, float64(attempt-1))
	if delay > float64(p.Max) {
		delay = float64(p.Max)
	}
	return time.Duration(delay)
}
func Run(ctx context.Context, p Policy, fn func(context.Context) error) error {
	if p.Attempts < 1 {
		p.Attempts = 1
	}
	var last error
	for attempt := 1; attempt <= p.Attempts; attempt++ {
		if err := fn(ctx); err == nil {
			return nil
		} else {
			last = err
		}
		if attempt == p.Attempts {
			break
		}
		timer := time.NewTimer(p.Delay(attempt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	return errors.Join(ErrExhausted, last)
}
func Permanent(err error) error { return errors.Join(ErrExhausted, err) }
