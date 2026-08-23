package clock

import (
	"testing"
	"time"
)

func TestClockCanAdvance(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := New(base)
	if !c.Now().Equal(base) {
		t.Fatal("initial")
	}
	c.Advance(time.Hour)
	if !c.UTC().Equal(base.Add(time.Hour)) {
		t.Fatal("advance")
	}
	c.Set(base)
	if !c.Now().Equal(base) {
		t.Fatal("set")
	}
}
