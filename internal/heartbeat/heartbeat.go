package heartbeat

import (
	"sync"
	"time"
)

type State struct {
	At      time.Time
	Healthy bool
	Message string
}
type Monitor struct {
	mu     sync.RWMutex
	states map[string]State
	ttl    time.Duration
}

func New(ttl time.Duration) *Monitor { return &Monitor{states: map[string]State{}, ttl: ttl} }
func (m *Monitor) Beat(name, message string, now time.Time) {
	m.mu.Lock()
	m.states[name] = State{At: now.UTC(), Healthy: true, Message: message}
	m.mu.Unlock()
}
func (m *Monitor) Healthy(name string, now time.Time) bool {
	m.mu.RLock()
	state, ok := m.states[name]
	m.mu.RUnlock()
	return ok && state.Healthy && now.Sub(state.At) <= m.ttl
}
func (m *Monitor) MarkUnhealthy(name, message string, now time.Time) {
	m.mu.Lock()
	m.states[name] = State{At: now.UTC(), Healthy: false, Message: message}
	m.mu.Unlock()
}
func (m *Monitor) Snapshot() map[string]State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := map[string]State{}
	for k, v := range m.states {
		out[k] = v
	}
	return out
}
func (m *Monitor) AllHealthy(now time.Time) bool {
	states := m.Snapshot()
	if len(states) == 0 {
		return false
	}
	for name := range states {
		if !m.Healthy(name, now) {
			return false
		}
	}
	return true
}
