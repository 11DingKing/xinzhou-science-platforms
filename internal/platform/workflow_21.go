package platform

import (
	"context"
)

func (s *Service) cancelAwareSubmit(ctx context.Context, actorID, platformID, version int64) error {
	_, err := s.db.Exec(`UPDATE innovation_platforms SET status='submitted',version=version+1 WHERE id=?`, platformID)
	return err
}
