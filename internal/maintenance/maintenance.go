package maintenance

import (
	"context"
	"sync"
	"time"
)

type Task struct {
	ID        string
	Run       func(context.Context) error
	Interval  time.Duration
	NextRun   time.Time
	Enabled   bool
	LastError string
	Runs      int
}
type Runner struct {
	mu    sync.Mutex
	tasks map[string]*Task
}

func New() *Runner { return &Runner{tasks: map[string]*Task{}} }
func (r *Runner) Add(task Task) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if task.ID == "" || task.Run == nil || task.Interval <= 0 {
		return false
	}
	if _, ok := r.tasks[task.ID]; ok {
		return false
	}
	r.tasks[task.ID] = &task
	return true
}
func (r *Runner) Tick(ctx context.Context, now time.Time) int {
	r.mu.Lock()
	tasks := []*Task{}
	for _, task := range r.tasks {
		if task.Enabled && !now.Before(task.NextRun) {
			// Claim the next run while holding the registry lock so concurrent
			// ticks cannot execute the same maintenance task twice.
			task.NextRun = now.Add(task.Interval)
			tasks = append(tasks, task)
		}
	}
	r.mu.Unlock()
	ran := 0
	for _, task := range tasks {
		err := task.Run(ctx)
		r.mu.Lock()
		task.Runs++
		if err != nil {
			task.LastError = err.Error()
		}
		r.mu.Unlock()
		ran++
	}
	return ran
}
func (r *Runner) Disable(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[id]
	if !ok {
		return false
	}
	task.Enabled = false
	return true
}
func (r *Runner) Snapshot() []Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []Task{}
	for _, task := range r.tasks {
		out = append(out, *task)
	}
	return out
}
