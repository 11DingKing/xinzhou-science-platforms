package limits

import (
	"testing"
)

func TestQuota(t *testing.T) {
	q := New(3)
	if err := q.Reserve(2); err != nil || q.Available() != 1 {
		t.Fatal("reserve")
	}
	if q.Reserve(2) != ErrExceeded {
		t.Fatal("over quota")
	}
	q.Release(1)
	if q.Used() != 1 {
		t.Fatal("release")
	}
	if q.Resize(0) == nil {
		t.Fatal("resize below used")
	}
	if q.Resize(5) != nil || q.Available() != 4 {
		t.Fatal("resize")
	}
}
