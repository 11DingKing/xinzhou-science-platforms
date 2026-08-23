package ownership

import (
	"testing"
	"time"
)

func TestOwnershipLease(t *testing.T) {
	now := time.Now()
	m := New()
	lease, err := m.Acquire("inspection:1", "reviewer", "token", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Acquire("inspection:1", "other", "other", now, time.Minute); err != ErrOwned {
		t.Fatal("ownership")
	}
	renew, err := m.Renew("inspection:1", "reviewer", "token", lease.Version, now, time.Minute)
	if err != nil || renew.Version != 2 {
		t.Fatal("renew")
	}
	if err := m.Release("inspection:1", "reviewer", "token", renew.Version); err != nil {
		t.Fatal(err)
	}
	if len(m.Expired(now.Add(time.Hour))) != 0 {
		t.Fatal("release")
	}
}
