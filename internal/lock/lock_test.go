package lock

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestKeyedSerializesSameKey(t *testing.T) {
	l := New()
	var mu sync.Mutex
	active := 0
	max := 0
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.With(context.Background(), "batch", func() error {
				mu.Lock()
				active++
				if active > max {
					max = active
				}
				mu.Unlock()
				time.Sleep(time.Millisecond)
				mu.Lock()
				active--
				mu.Unlock()
				return nil
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if max != 1 {
		t.Fatalf("max active=%d", max)
	}
}
