package notification

import (
	"context"
	"encoding/json"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"time"
)

type Service struct{ repos *repository.Repos }

func New(r *repository.Repos) *Service { return &Service{repos: r} }
func (s *Service) Queue(ctx context.Context, recipientID int64, kind string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.repos.EnqueueNotification(ctx, recipientID, kind, string(b), time.Now())
}
func (s *Service) Deliver(ctx context.Context, limit int) (int, error) {
	return s.repos.DeliverNotifications(ctx, limit, time.Now())
}
