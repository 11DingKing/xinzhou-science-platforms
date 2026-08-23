package ownership

import (
	"errors"
	"sync"
	"time"
)

var ErrOwned = errors.New("resource already owned")

type Lease struct {
	Resource  string
	Owner     string
	Token     string
	ExpiresAt time.Time
	Version   int64
}
type Manager struct {
	mu     sync.Mutex
	leases map[string]Lease
}

func New() *Manager { return &Manager{leases: map[string]Lease{}} }
func (m *Manager) Acquire(resource, owner, token string, now time.Time, ttl time.Duration) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if resource == "" || owner == "" || token == "" {
		return Lease{}, errors.New("lease fields required")
	}
	if old, ok := m.leases[resource]; ok && now.Before(old.ExpiresAt) {
		return Lease{}, ErrOwned
	}
	lease := Lease{Resource: resource, Owner: owner, Token: token, ExpiresAt: now.Add(ttl), Version: 1}
	if old, ok := m.leases[resource]; ok {
		lease.Version = old.Version + 1
	}
	m.leases[resource] = lease
	return lease, nil
}
func (m *Manager) Renew(resource, owner, token string, version int64, now time.Time, ttl time.Duration) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.leases[resource]
	if !ok || old.Owner != owner || old.Token != token || old.Version != version || !now.Before(old.ExpiresAt) {
		return Lease{}, errors.New("lease cannot renew")
	}
	old.ExpiresAt = now.Add(ttl)
	old.Version++
	m.leases[resource] = old
	return old, nil
}
func (m *Manager) Release(resource, owner, token string, version int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.leases[resource]
	if !ok || old.Owner != owner || old.Token != token || old.Version != version {
		return errors.New("lease cannot release")
	}
	delete(m.leases, resource)
	return nil
}
func (m *Manager) Expired(now time.Time) []Lease {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Lease{}
	for _, lease := range m.leases {
		if !now.Before(lease.ExpiresAt) {
			out = append(out, lease)
		}
	}
	return out
}
