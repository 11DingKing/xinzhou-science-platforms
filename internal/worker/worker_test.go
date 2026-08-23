package worker

import (
	"context"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkerProcessesQueuedJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "worker.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := repository.New(db)
	u, err := r.CreateUser(ctx, "worker-user", "p", domain.RoleReviewer)
	if err != nil {
		t.Fatal(err)
	}
	_ = u
	if err := r.Enqueue(ctx, "sample", "{}"); err != nil {
		t.Fatal(err)
	}
	w := New(r)
	go w.Run(ctx)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		err := db.QueryRowContext(ctx, `SELECT status FROM jobs LIMIT 1`).Scan(&status)
		if err == nil && status == "done" {
			w.Stop()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	w.Stop()
	t.Fatal("worker did not complete job")
}
func TestWorkerStopsOnContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "stop.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	w := New(repository.New(db))
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}
