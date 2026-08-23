package fulfillment

import (
	"context"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"time"
)

type Rule struct {
	ID                         int64
	VersionID                  int64
	Region                     string
	EffectiveFrom, EffectiveTo time.Time
	Status                     string
	Version                    int64
}
type Service struct{ repos *repository.Repos }

func New(r *repository.Repos) *Service { return &Service{repos: r} }
func (s *Service) Draft(ctx context.Context, actor domain.User, versionID int64, region string, from, to time.Time) (Rule, error) {
	if actor.Role != domain.RoleMerchant {
		return Rule{}, apperrors.ErrForbidden
	}
	if region == "" || !to.After(from) {
		return Rule{}, errors.New("invalid effective window")
	}
	id, err := s.repos.CreateFulfillmentRule(ctx, actor.ID, versionID, region, from, to)
	return Rule{ID: id, VersionID: versionID, Region: region, EffectiveFrom: from, EffectiveTo: to, Status: "draft", Version: 1}, err
}
func (s *Service) Publish(ctx context.Context, actor domain.User, id, version int64) error {
	return s.repos.PublishFulfillmentRule(ctx, actor.ID, id, version)
}
