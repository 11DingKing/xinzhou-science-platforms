package remediation

import (
	"context"
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
func (s *Service) Plan(ctx context.Context, actor domain.User, complaintID int64, action string, due time.Time) (domain.Remediation, error) {
	if actor.Role != domain.RoleReviewer {
		return domain.Remediation{}, apperrors.ErrForbidden
	}
	return s.repos.CreateRemediation(ctx, actor.ID, complaintID, action, due)
}
func (s *Service) Activate(ctx context.Context, actor domain.User, id, version int64) error {
	return s.repos.TransitionRemediation(ctx, actor.ID, id, version, domain.RemediationActive)
}
func (s *Service) Complete(ctx context.Context, actor domain.User, id, version int64) error {
	return s.repos.TransitionRemediation(ctx, actor.ID, id, version, domain.RemediationDone)
}
func (s *Service) EscalateExpired(ctx context.Context) (int, error) {
	return s.repos.EscalateRemediation(ctx, s.now())
}
