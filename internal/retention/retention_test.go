package retention

import (
	"testing"
	"time"
)

func TestPolicyCutoffs(t *testing.T) {
	p := DefaultPolicy()
	now := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
	a, e, s := p.Cutoffs(now)
	if a.Year() != 2025 || e.Year() != 2026 || s.Year() != 2026 {
		t.Fatalf("cutoffs=%v %v %v", a, e, s)
	}
	if !ShouldArchive(now.AddDate(0, 0, -181), now, 180) || ShouldArchive(now, now, 180) {
		t.Fatal("archive decision")
	}
}
