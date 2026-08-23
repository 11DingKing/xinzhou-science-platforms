package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrConflict = errors.New("idempotency key reused with different request")

type Entry struct {
	Key       string
	Hash      string
	Response  []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}
type Store struct {
	mu    sync.Mutex
	items map[string]Entry
}

func New() *Store                { return &Store{items: map[string]Entry{}} }
func Hash(payload []byte) string { sum := sha256.Sum256(payload); return hex.EncodeToString(sum[:]) }
func (s *Store) Put(key string, payload, response []byte, now time.Time, ttl time.Duration) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := Hash(payload)
	if old, ok := s.items[key]; ok && now.Before(old.ExpiresAt) {
		if old.Hash != hash {
			return nil, false, ErrConflict
		}
		return append([]byte(nil), old.Response...), true, nil
	}
	s.items[key] = Entry{Key: key, Hash: hash, Response: append([]byte(nil), response...), CreatedAt: now, ExpiresAt: now.Add(ttl)}
	return nil, false, nil
}
func (s *Store) Purge(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for key, value := range s.items {
		if !now.Before(value.ExpiresAt) {
			delete(s.items, key)
			n++
		}
	}
	return n
}
func (s *Store) Len() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.items) }
