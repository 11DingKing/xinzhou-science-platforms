package lifecycle

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type State string

const (
	StateNew      State = "new"
	StateRunning  State = "running"
	StateStopping State = "stopping"
	StateStopped  State = "stopped"
	StateFailed   State = "failed"
)

type Manager struct {
	mu      sync.Mutex
	state   State
	changed time.Time
	reason  string
}

func New(now time.Time) *Manager { return &Manager{state: StateNew, changed: now.UTC()} }
func (m *Manager) Transition(to State, now time.Time, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !allowed(m.state, to) {
		return fmt.Errorf("invalid lifecycle %s -> %s", m.state, to)
	}
	m.state = to
	m.changed = now.UTC()
	m.reason = reason
	return nil
}
func allowed(from, to State) bool {
	switch from {
	case StateNew:
		return to == StateRunning || to == StateFailed
	case StateRunning:
		return to == StateStopping || to == StateFailed
	case StateStopping:
		return to == StateStopped || to == StateFailed
	case StateFailed:
		return to == StateRunning
	default:
		return false
	}
}
func (m *Manager) Snapshot() (State, time.Time, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state, m.changed, m.reason
}
func (m *Manager) RequireRunning() error {
	state, _, _ := m.Snapshot()
	if state != StateRunning {
		return errors.New("manager is not running")
	}
	return nil
}
func (m *Manager) Uptime(now time.Time) time.Duration {
	_, changed, _ := m.Snapshot()
	if now.Before(changed) {
		return 0
	}
	return now.Sub(changed)
}
