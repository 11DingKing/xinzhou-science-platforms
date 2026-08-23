package platform

import (
	"context"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
)

func (s *Service) versionSafeMilestone(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE platform_milestones SET status='complete',version=version+1,completed_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), platformID); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `UPDATE innovation_platforms SET version=version+1 WHERE id=? AND version=?`, platformID, version-1)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return apperrors.ErrConflict
	}
	return nil
}
