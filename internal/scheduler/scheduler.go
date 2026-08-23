package scheduler

import (
	"context"
	"sort"
	"sync"
	"time"
)

type Job struct {
	ID       string
	RunAt    time.Time
	Payload  any
	Attempts int
	Status   string
}
type Scheduler struct {
	mu   sync.Mutex
	jobs map[string]Job
	wake chan struct{}
}

func New() *Scheduler { return &Scheduler{jobs: map[string]Job{}, wake: make(chan struct{}, 1)} }
func (s *Scheduler) Schedule(job Job) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job.ID == "" || job.RunAt.IsZero() || job.Status == "" {
		return false
	}
	if _, exists := s.jobs[job.ID]; exists {
		return false
	}
	s.jobs[job.ID] = job
	select {
	case s.wake <- struct{}{}:
	default:
	}
	return true
}
func (s *Scheduler) Due(now time.Time) []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Job{}
	for _, job := range s.jobs {
		if job.Status == "queued" && !now.Before(job.RunAt) {
			out = append(out, job)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RunAt.Before(out[j].RunAt) })
	return out
}
func (s *Scheduler) Mark(id, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return false
	}
	job.Status = status
	job.Attempts++
	s.jobs[id] = job
	return true
}
func (s *Scheduler) Run(ctx context.Context, handler func(context.Context, Job) error) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.wake:
		case <-ticker.C:
		}
		for _, job := range s.Due(time.Now()) {
			s.Mark(job.ID, "running")
			if err := handler(ctx, job); err != nil {
				s.Mark(job.ID, "failed")
			} else {
				s.Mark(job.ID, "done")
			}
		}
	}
}
