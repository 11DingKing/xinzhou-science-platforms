package batch

import (
	"context"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"time"
)

type Service struct {
	repos *repository.Repos
	now   func() time.Time
}

func New(r *repository.Repos) *Service { return &Service{repos: r, now: time.Now} }
func (s *Service) Register(ctx context.Context, actor domain.User, versionID int64, code, region string, expires time.Time) (domain.Batch, error) {
	if actor.Role != domain.RoleMerchant {
		return domain.Batch{}, apperrors.ErrForbidden
	}
	if code == "" || region == "" || expires.Before(s.now()) {
		return domain.Batch{}, errors.New("invalid batch details")
	}
	return s.repos.CreateBatch(ctx, actor, domain.Batch{VersionID: versionID, Code: code, Region: region, ExpiresAt: expires})
}
func (s *Service) StartSampling(ctx context.Context, actor domain.User, batchID int64, version int64) error {
	return s.repos.UpdateBatchStatus(ctx, actor, batchID, version, domain.BatchSampling)
}
func (s *Service) Finalize(ctx context.Context, actor domain.User, batchID, version int64, flagged bool) error {
	status := domain.BatchCleared
	if flagged {
		status = domain.BatchFlagged
	}
	return s.repos.UpdateBatchStatus(ctx, actor, batchID, version, status)
}
