package investigation

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"strings"
	"testing"
	"time"
)

func TestInvestigationLifecycle(t *testing.T) {
	now := time.Now()
	c, err := Start(Case{ID: 1, Status: domain.CaseOpen}, 9, now)
	if err != nil || c.OwnerID != 9 {
		t.Fatal("start")
	}
	if err := AddFinding(&c, Finding{ID: 1, Severity: "high", Statement: "material differs", EvidenceIDs: []int64{2}}); err != nil {
		t.Fatal(err)
	}
	f := Finding{ID: 1, Severity: "high", Statement: "material differs", EvidenceIDs: []int64{2}}
	if err := AddFinding(&c, f); err != nil {
		t.Fatal(err)
	}
	if err := AddFinding(&c, Finding{ID: 2, Severity: "critical", Statement: "region split", EvidenceIDs: []int64{3}}); err != nil {
		t.Fatal(err)
	}
	c, err = Resolve(c, now)
	if err != nil {
		t.Fatal(err)
	}
	c, err = Close(c, now)
	if err != nil || c.ClosedAt == nil {
		t.Fatal("close")
	}
	if Severity(c.Findings) != "critical" || !strings.Contains(Explain(c), "status=closed") {
		t.Fatal("summary")
	}
}
func TestAssignmentAndSorting(t *testing.T) {
	now := time.Now()
	c := Case{ID: 2, Status: domain.CaseInvestigating}
	a, err := Assign(c, 4, now, time.Hour)
	if err != nil || !a.DueAt.After(a.AssignedAt) {
		t.Fatal("assignment")
	}
	sorted := SortFindings([]Finding{{ID: 2, Severity: "low"}, {ID: 1, Severity: "critical"}})
	if sorted[0].ID != 1 {
		t.Fatal("sort")
	}
}
