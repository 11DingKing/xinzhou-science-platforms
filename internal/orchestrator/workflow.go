package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"time"
)

type Workflow struct {
	repos *repository.Repos
	clock func() time.Time
}

func New(r *repository.Repos) *Workflow { return &Workflow{repos: r, clock: time.Now} }

type ReleaseInput struct {
	Version    domain.ProductVersion
	Batch      domain.Batch
	ReviewerID int64
	RequestID  string
}
type ReleaseResult struct {
	VersionID  int64
	BatchID    int64
	AuditCount int
	QueuedJob  bool
	ReleasedAt time.Time
}

func (w *Workflow) Release(ctx context.Context, actor domain.User, in ReleaseInput) (ReleaseResult, error) {
	if actor.Role != domain.RoleMerchant {
		return ReleaseResult{}, apperrors.ErrForbidden
	}
	if in.Version.ID < 1 || in.Batch.ID < 1 {
		return ReleaseResult{}, errors.New("release requires persisted objects")
	}
	if in.Version.Status != domain.VersionPublished {
		return ReleaseResult{}, apperrors.ErrInvalidState
	}
	if in.Batch.VersionID != in.Version.ID {
		return ReleaseResult{}, errors.New("batch version mismatch")
	}
	if in.ReviewerID < 1 {
		return ReleaseResult{}, errors.New("reviewer required")
	}
	if err := w.repos.UpdateBatchStatus(ctx, actor, in.Batch.ID, in.Batch.Version, domain.BatchSampling); err != nil {
		return ReleaseResult{}, err
	}
	if err := w.repos.Enqueue(ctx, "inspection", fmt.Sprintf(`{"batch_id":%d,"reviewer_id":%d}`, in.Batch.ID, in.ReviewerID)); err != nil {
		return ReleaseResult{}, err
	}
	return ReleaseResult{VersionID: in.Version.ID, BatchID: in.Batch.ID, QueuedJob: true, ReleasedAt: w.clock().UTC()}, nil
}
func (w *Workflow) Flag(ctx context.Context, actor domain.User, batchID, version int64, reason string) error {
	if reason == "" {
		return errors.New("flag reason required")
	}
	return w.repos.UpdateBatchStatus(ctx, actor, batchID, version, domain.BatchFlagged)
}
func (w *Workflow) Archive(ctx context.Context, actor domain.User, batchID, version int64) error {
	return w.repos.UpdateBatchStatus(ctx, actor, batchID, version, domain.BatchArchived)
}
func AllowedRelease(actor domain.User) bool {
	return actor.Role == domain.RoleMerchant && !actor.Disabled
}
