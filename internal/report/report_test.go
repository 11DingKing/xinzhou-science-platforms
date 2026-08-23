package report

import (
	"strings"
	"testing"
	"time"
)

func TestReportQueriesCarrySameFilters(t *testing.T) {
	f := Filter{Search: "lamp", Region: "county", Status: "flagged", From: time.Now().Add(-time.Hour), To: time.Now()}
	a, args := BuildBatchQuery(f)
	b, bargs := BuildComplaintQuery(f)
	if !strings.Contains(a, "b.region=?") || !strings.Contains(b, "c.region=?") || len(args) != 9 || len(bargs) != 7 {
		t.Fatalf("queries args=%d,%d", len(args), len(bargs))
	}
	if ValidateFilter(Filter{Limit: 101}) == nil {
		t.Fatal("limit")
	}
	if len(StatusCounts([]ComplaintRow{{Status: "open"}, {Status: "closed"}})) != 2 {
		t.Fatal("counts")
	}
}
