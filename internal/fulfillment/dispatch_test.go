package fulfillment

import (
	"testing"
	"time"
)

func TestDispatchResolution(t *testing.T) {
	now := time.Now()
	s := RuleSet{DefaultTier: "standard", Rules: []Rule{{ID: 2, Region: "county", EffectiveFrom: now.Add(-time.Hour), EffectiveTo: now.Add(time.Hour), Status: "published"}}}
	d := s.Resolve(DispatchRequest{Region: "County"}, now)
	if !d.Allowed || d.RuleID != 2 || d.QualityTier != "declared" {
		t.Fatalf("decision=%+v", d)
	}
	if s.Resolve(DispatchRequest{Region: "city"}, now).QualityTier != "standard" {
		t.Fatal("default")
	}
	if s.Validate() != nil {
		t.Fatal("valid rules")
	}
}
func TestDispatchConsistency(t *testing.T) {
	if !CompareRequests(DispatchRequest{SKU: "S", BatchCode: "B", Region: "city"}, DispatchRequest{SKU: "s", BatchCode: "b", Region: "CITY"}) {
		t.Fatal("case comparison")
	}
	if RequireConsistentTier([]DispatchDecision{{QualityTier: "a"}, {QualityTier: "b"}}) == nil {
		t.Fatal("inconsistent tiers")
	}
}
