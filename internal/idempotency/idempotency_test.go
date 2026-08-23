package idempotency

import (
	"testing"
	"time"
)

func TestIdempotencyStore(t *testing.T) {
	now := time.Now()
	s := New()
	if _, replayed, err := s.Put("k", []byte("a"), []byte("response"), now, time.Minute); err != nil || replayed {
		t.Fatal("first")
	}
	got, replayed, err := s.Put("k", []byte("a"), []byte("other"), now.Add(time.Second), time.Minute)
	if err != nil || !replayed || string(got) != "response" {
		t.Fatal("replay")
	}
	if _, _, err := s.Put("k", []byte("b"), []byte("x"), now.Add(time.Second), time.Minute); err != ErrConflict {
		t.Fatal("conflict")
	}
	if s.Purge(now.Add(time.Minute)) != 1 || s.Len() != 0 {
		t.Fatal("purge")
	}
}
