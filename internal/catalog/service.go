package catalog

import (
	"context"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
)

type Service struct{ repos *repository.Repos }

func New(r *repository.Repos) *Service { return &Service{repos: r} }
func (s *Service) Register(ctx context.Context, actor domain.User, sku, name, channel string) (domain.ProductVersion, error) {
	if actor.Role != domain.RoleMerchant {
		return domain.ProductVersion{}, apperrors.ErrForbidden
	}
	if sku == "" || name == "" || channel == "" {
		return domain.ProductVersion{}, errors.New("sku, name and channel are required")
	}
	return s.repos.CreateVersion(ctx, actor, domain.ProductVersion{SKU: sku, DisplayName: name, Channel: channel})
}
func (s *Service) Publish(ctx context.Context, actor domain.User, id, version int64) error {
	if actor.Role != domain.RoleMerchant {
		return apperrors.ErrForbidden
	}
	return s.repos.PublishVersion(ctx, actor, id, version)
}
