package worker

import (
	"context"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"time"
)

type Worker struct {
	repos *repository.Repos
	stop  chan struct{}
}

func New(r *repository.Repos) *Worker { return &Worker{repos: r, stop: make(chan struct{})} }
func (w *Worker) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
}
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}
func (w *Worker) tick(ctx context.Context) {
	id, err := w.repos.ClaimJob(ctx, "quality-worker", 2*time.Second)
	if err != nil {
		return
	}
	if err := w.repos.CompleteJob(ctx, id); err != nil {
		_ = w.repos.FailJob(ctx, id, err.Error())
	}
}
