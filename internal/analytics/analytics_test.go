package analytics

import (
	"testing"
	"time"
)

func TestBuildRegionSummary(t *testing.T) {
	now := time.Now()
	s := Build([]Observation{{Region: "city", Score: 90}, {Region: "city", Score: 80}, {Region: "county", Score: 40}}, now)
	if len(s.ByRegion) != 2 || s.FlaggedRegions != 1 || s.TotalSamples != 3 {
		t.Fatalf("summary=%+v", s)
	}
	if Worst(s, 1)[0].Region != "county" {
		t.Fatal("worst")
	}
	if Compare(s, s)["city"] != 0 {
		t.Fatal("compare")
	}
}
