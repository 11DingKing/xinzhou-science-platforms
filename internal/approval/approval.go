package approval

import (
	"errors"
	"fmt"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"time"
)

var ErrNotEligible = errors.New("approval prerequisites incomplete")

type Checklist struct {
	HasDeclaration, HasInspection, HasEvidence, HasReviewerSignoff bool
	SubmittedAt                                                    time.Time
}
type Result struct {
	Approved   bool
	Reason     string
	ApprovedAt time.Time
}

func Evaluate(c Checklist, now time.Time) Result {
	if !c.HasDeclaration {
		return Result{Reason: "declaration missing"}
	}
	if !c.HasInspection {
		return Result{Reason: "inspection missing"}
	}
	if !c.HasEvidence {
		return Result{Reason: "evidence missing"}
	}
	if !c.HasReviewerSignoff {
		return Result{Reason: "reviewer signoff missing"}
	}
	if c.SubmittedAt.IsZero() || c.SubmittedAt.After(now) {
		return Result{Reason: "submission time invalid"}
	}
	return Result{Approved: true, ApprovedAt: now.UTC()}
}
func RequireForBatch(b domain.Batch, c Checklist, now time.Time) error {
	if b.Status != domain.BatchSampling {
		return fmt.Errorf("batch %d is not reviewable", b.ID)
	}
	if !Evaluate(c, now).Approved {
		return ErrNotEligible
	}
	return nil
}
func Expired(result Result, now time.Time, ttl time.Duration) bool {
	return result.Approved && result.ApprovedAt.Add(ttl).Before(now)
}
