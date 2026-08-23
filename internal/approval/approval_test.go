package approval

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"testing"
	"time"
)

func TestApprovalChecklist(t *testing.T) {
	now := time.Now()
	base := Checklist{HasDeclaration: true, HasInspection: true, HasEvidence: true, HasReviewerSignoff: true, SubmittedAt: now}
	if !Evaluate(base, now).Approved {
		t.Fatal("valid checklist denied")
	}
	base.HasEvidence = false
	if Evaluate(base, now).Approved {
		t.Fatal("missing evidence approved")
	}
	if RequireForBatch(domain.Batch{ID: 1, Status: domain.BatchSampling}, Checklist{}, now) == nil {
		t.Fatal("incomplete batch approved")
	}
	r := Evaluate(Checklist{HasDeclaration: true, HasInspection: true, HasEvidence: true, HasReviewerSignoff: true, SubmittedAt: now}, now)
	if !Expired(r, now.Add(time.Hour), 30*time.Minute) {
		t.Fatal("expiry")
	}
}
