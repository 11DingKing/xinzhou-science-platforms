package scoring

import (
	"strings"
	"testing"
)

func TestScoringEngine(t *testing.T) {
	e := DefaultEngine()
	r := e.Evaluate(Input{MaterialMatch: false, ConfigurationMatch: false, RegionMatch: false, ComplaintCount: 2, InspectionScore: 40})
	if r.Score != 100 || r.Level != "critical" || len(r.Reasons) != 5 {
		t.Fatalf("result=%+v", r)
	}
	if !strings.Contains(Explain(r), "material") {
		t.Fatal("explain")
	}
	if e.Validate() != nil {
		t.Fatal("rules")
	}
	if SortRules(e.Rules)[0].Weight != 30 {
		t.Fatal("sort")
	}
}
