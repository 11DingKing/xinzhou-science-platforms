package ledger

import (
	"testing"
	"time"
)

func TestLedgerOrdering(t *testing.T) {
	l := New()
	now := time.Now()
	if _, err := l.Append("batch", 1, 2, "created", "r1", now); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Append("batch", 1, 2, "flagged", "r2", now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if len(l.List("batch", 1)) != 2 || len(l.Since(now.Add(time.Second))) != 1 || l.Count() != 2 {
		t.Fatal("ledger")
	}
	if _, err := l.Append("", 0, 0, "", "", now); err == nil {
		t.Fatal("invalid")
	}
}
