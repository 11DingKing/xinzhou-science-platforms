package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestSchedulerDeduplicatesAndRuns(t *testing.T) {
	s := New()
	now := time.Now()
	job := Job{ID: "one", RunAt: now.Add(-time.Second), Payload: "x", Status: "queued"}
	if !s.Schedule(job) || s.Schedule(job) {
		t.Fatal("duplicate")
	}
	if len(s.Due(now)) != 1 {
		t.Fatal("due")
	}
	if !s.Mark("one", "done") || s.Mark("missing", "done") {
		t.Fatal("mark")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Schedule(Job{ID: "two", RunAt: now.Add(-time.Second), Status: "queued"})
	done := make(chan struct{})
	go func() { s.Run(ctx, func(context.Context, Job) error { close(done); return nil }) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("not run")
	}
}
