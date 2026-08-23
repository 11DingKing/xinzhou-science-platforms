package sampling

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"testing"
	"time"
)

func TestPlanLifecycle(t *testing.T) {
	now := time.Now()
	p, err := BuildPlan(4, 7, []RegionQuota{{Region: "County", Quota: 2, Priority: 1}, {Region: "city", Quota: 1, Priority: 2}}, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if p.Regions[0].Region != "city" {
		t.Fatal("priority")
	}
	p, err = Publish(p, now)
	if err != nil || p.Status != "published" {
		t.Fatalf("publish=%+v err=%v", p, err)
	}
	selected := SelectCandidates(p, []Candidate{{Region: "city"}, {Region: "county"}, {Region: "county"}, {Region: "county"}})
	if len(selected) != 3 {
		t.Fatalf("selected=%d", len(selected))
	}
	if ValidatePlan(p, domain.Batch{ID: 7, Status: domain.BatchSampling}, now) != nil {
		t.Fatal("validation")
	}
}
func TestPlanRejectsInvalidInput(t *testing.T) {
	now := time.Now()
	if _, err := BuildPlan(1, 1, nil, now, time.Hour); err == nil {
		t.Fatal("empty regions")
	}
	if _, err := BuildPlan(1, 1, []RegionQuota{{Region: "x", Quota: 0}}, now, time.Hour); err == nil {
		t.Fatal("zero quota")
	}
}
