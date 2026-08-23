package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
)

type Service struct{ repos *repository.Repos }

func New(r *repository.Repos) *Service { return &Service{repos: r} }
func (s *Service) Attach(ctx context.Context, actorID, complaintID int64, objectKey string, payload []byte) (string, error) {
	if objectKey == "" || len(payload) == 0 {
		return "", errors.New("evidence payload is required")
	}
	sum := sha256.Sum256(payload)
	hash := hex.EncodeToString(sum[:])
	if err := s.repos.AttachEvidence(ctx, actorID, complaintID, objectKey, hash); err != nil {
		return "", err
	}
	return hash, nil
}
func (s *Service) Archive(ctx context.Context, actorID, evidenceID int64) error {
	if err := s.repos.ArchiveEvidence(ctx, actorID, evidenceID); err != nil {
		return err
	}
	return nil
}
func RequireArchived(ok bool) error {
	if !ok {
		return apperrors.ErrInvalidState
	}
	return nil
}
