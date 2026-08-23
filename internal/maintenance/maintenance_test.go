package maintenance

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMaintenanceRunner(t *testing.T) {
	r := New()
	now := time.Now()
	calls := 0
	if !r.Add(Task{ID: "purge", Run: func(context.Context) error { calls++; return nil }, Interval: time.Hour, NextRun: now, Enabled: true}) {
		t.Fatal("add")
	}
	if r.Add(Task{ID: "purge", Run: func(context.Context) error { return nil }, Interval: time.Hour}) {
		t.Fatal("duplicate")
	}
	if r.Tick(context.Background(), now) != 1 || calls != 1 {
		t.Fatal("tick")
	}
	if !r.Disable("purge") || r.Tick(context.Background(), now.Add(2*time.Hour)) != 0 {
		t.Fatal("disable")
	}
	if !r.Add(Task{ID: "fail", Run: func(context.Context) error { return errors.New("bad") }, Interval: time.Hour, NextRun: now, Enabled: true}) {
		t.Fatal("fail add")
	}
	r.Tick(context.Background(), now)
	found := false
	for _, task := range r.Snapshot() {
		if task.ID == "fail" && task.LastError != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("error not stored")
	}
}

func TestConcurrentTickClaimsMaintenanceTaskOnce(t *testing.T) {
	r := New()
	started := make(chan struct{})
	release := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	if !r.Add(Task{ID: "audit", Run: func(context.Context) error {
		mu.Lock()
		calls++
		mu.Unlock()
		close(started)
		<-release
		return nil
	}, Interval: time.Hour, NextRun: time.Now(), Enabled: true}) {
		t.Fatal("add failed")
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); r.Tick(context.Background(), time.Now()) }()
	<-started
	go func() { defer wg.Done(); r.Tick(context.Background(), time.Now()) }()
	time.Sleep(20 * time.Millisecond)
	close(release)
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("maintenance task ran %d times", calls)
	}
}
