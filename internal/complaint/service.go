package complaint

import (
	"context"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
)

type Service struct{ repos *repository.Repos }

func New(r *repository.Repos) *Service { return &Service{repos: r} }
func (s *Service) Open(ctx context.Context, reporter domain.User, versionID, batchID int64, region, description, key string) (domain.Complaint, error) {
	if description == "" || region == "" {
		return domain.Complaint{}, errors.New("description and region are required")
	}
	if reporter.Disabled {
		return domain.Complaint{}, apperrors.ErrForbidden
	}
	return s.repos.OpenComplaint(ctx, reporter.ID, versionID, batchID, region, description, key)
}
func (s *Service) BeginInvestigation(ctx context.Context, actor domain.User, id, version int64) error {
	return s.repos.TransitionComplaint(ctx, actor.ID, id, version, domain.CaseInvestigating)
}
func (s *Service) Resolve(ctx context.Context, actor domain.User, id, version int64) error {
	return s.repos.TransitionComplaint(ctx, actor.ID, id, version, domain.CaseResolved)
}
func (s *Service) Close(ctx context.Context, actor domain.User, id, version int64) error {
	return s.repos.CloseComplaintIfEvidenceArchived(ctx, actor.ID, id, version)
}
