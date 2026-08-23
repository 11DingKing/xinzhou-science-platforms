package appeal

import (
	"testing"
	"time"
)

func TestAppealLifecycle(t *testing.T) {
	now := time.Now()
	a, err := Submit(Appeal{ID: 1, Status: StatusDraft, Reason: "evidence from county shipment", EvidenceIDs: []int64{1}}, now)
	if err != nil || a.Status != StatusSubmitted {
		t.Fatal("submit")
	}
	a, err = StartReview(a, 8)
	if err != nil {
		t.Fatal(err)
	}
	a, err = Decide(a, false)
	if err != nil || !IsFinal(a.Status) || CanEdit(a) {
		t.Fatal("decide")
	}
	if ReasonWords(a) != 4 {
		t.Fatal("words")
	}
}
func TestAppealNeedsEvidence(t *testing.T) {
	if _, err := Submit(Appeal{Status: StatusDraft, Reason: "reason"}, time.Now()); err == nil {
		t.Fatal("empty evidence")
	}
}
