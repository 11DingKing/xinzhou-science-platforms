package operation

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

type Operation struct {
	ID                    string
	Kind                  string
	State                 State
	OwnerID               int64
	Progress              int
	Message               string
	StartedAt, FinishedAt *time.Time
	Version               int64
}
type Store struct {
	mu    sync.Mutex
	items map[string]Operation
}

func New() *Store { return &Store{items: map[string]Operation{}} }
func (s *Store) Create(id, kind string, owner int64) (Operation, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(kind) == "" || owner < 1 {
		return Operation{}, errors.New("invalid operation")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; ok {
		return Operation{}, errors.New("operation exists")
	}
	op := Operation{ID: id, Kind: kind, OwnerID: owner, State: StateQueued, Version: 1}
	s.items[id] = op
	return op, nil
}
func (s *Store) Start(id string, now time.Time) (Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.items[id]
	if !ok || op.State != StateQueued {
		return Operation{}, errors.New("operation not queued")
	}
	op.State = StateRunning
	op.Progress = 1
	op.Version++
	t := now.UTC()
	op.StartedAt = &t
	s.items[id] = op
	return op, nil
}
func (s *Store) Update(id string, progress int, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.items[id]
	if !ok || op.State != StateRunning {
		return errors.New("operation not running")
	}
	if progress < op.Progress || progress > 100 {
		return errors.New("invalid progress")
	}
	op.Progress = progress
	op.Message = message
	op.Version++
	s.items[id] = op
	return nil
}
func (s *Store) Finish(id string, state State, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.items[id]
	if !ok || op.State != StateRunning {
		return errors.New("operation not running")
	}
	if state != StateSucceeded && state != StateFailed && state != StateCancelled {
		return errors.New("invalid final state")
	}
	op.State = state
	op.Progress = 100
	op.Version++
	t := now.UTC()
	op.FinishedAt = &t
	s.items[id] = op
	return nil
}
func (s *Store) Get(id string) (Operation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.items[id]
	return op, ok
}
