package inspection

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
	lease time.Duration
}

func New(r *repository.Repos) *Service { return &Service{repos: r, lease: 5 * time.Minute} }
func (s *Service) Claim(ctx context.Context, reviewer domain.User) (domain.Inspection, error) {
	if reviewer.Role != domain.RoleReviewer {
		return domain.Inspection{}, apperrors.ErrForbidden
	}
	return s.repos.ClaimInspection(ctx, reviewer.ID, s.lease)
}
func (s *Service) Submit(ctx context.Context, reviewer domain.User, id, version int64, result, notes string) error {
	if result != "pass" && result != "fail" {
		return errors.New("result must be pass or fail")
	}
	return s.repos.SubmitInspection(ctx, reviewer.ID, id, version, result, notes)
}
func (s *Service) ReopenExpired(ctx context.Context) (int, error) {
	return s.repos.RequeueExpiredInspections(ctx, time.Now())
}
