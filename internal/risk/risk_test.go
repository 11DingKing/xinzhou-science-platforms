package risk

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"testing"
	"time"
)

func TestRiskScoreAndRanking(t *testing.T) {
	c := NewCalculator()
	now := time.Now()
	a := c.Score([]Signal{{Kind: "material_mismatch", BatchID: 2, ObservedAt: now}, {Kind: "configuration_mismatch", BatchID: 2, ObservedAt: now}})
	b := c.Score([]Signal{{Kind: "region_mismatch", BatchID: 1, ObservedAt: now}})
	if a.Score != 70 || a.Level != "high" || !IsActionable(a) {
		t.Fatalf("a=%+v", a)
	}
	if Rank([]Profile{b, a})[0].BatchID != 2 {
		t.Fatal("ranking")
	}
	if len(Explain(a)) != 2 {
		t.Fatal("explain")
	}
	if !SupportsBatch(a, domain.Batch{ID: 2, Status: domain.BatchFlagged}) {
		t.Fatal("support")
	}
}
func TestRiskBounds(t *testing.T) {
	c := NewCalculator()
	signals := []Signal{}
	for i := 0; i < 10; i++ {
		signals = append(signals, Signal{Kind: "material_mismatch", BatchID: 4})
	}
	p := c.Score(signals)
	if p.Score != 100 || p.Level != "critical" {
		t.Fatalf("profile=%+v", p)
	}
}
