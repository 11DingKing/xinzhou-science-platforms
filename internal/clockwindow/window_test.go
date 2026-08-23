package clockwindow

import (
	"testing"
	"time"
)

func TestWindowTimezone(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	start := time.Date(2026, 1, 1, 9, 0, 0, 0, loc)
	w, err := New(start, start.Add(2*time.Hour), loc)
	if err != nil || !w.Contains(start.Add(time.Hour)) {
		t.Fatal("contains")
	}
	if w.Remaining(start) != 2*time.Hour || len(Split(w, 2)) != 2 {
		t.Fatal("window")
	}
	if _, err := New(start, start.Add(-time.Hour), loc); err == nil {
		t.Fatal("reversed")
	}
}
