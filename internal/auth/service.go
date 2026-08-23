package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"time"
)

type Service struct{ repos *repository.Repos }

func NewService(r *repository.Repos) *Service { return &Service{repos: r} }
func (s *Service) Login(ctx context.Context, email, pass string) (string, domain.User, error) {
	u, err := s.repos.FindUser(ctx, email)
	if err != nil || u.PasswordHash != pass || u.Disabled {
		return "", domain.User{}, apperrors.ErrUnauthorized
	}
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", u, err
	}
	token := hex.EncodeToString(b)
	if err := s.repos.CreateSession(ctx, token, u.ID, time.Now().Add(time.Hour)); err != nil {
		return "", u, err
	}
	return token, u, nil
}
func (s *Service) Authenticate(ctx context.Context, token string) (domain.User, error) {
	return s.repos.GetSession(ctx, token)
}
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repos.RevokeSession(ctx, token)
}
