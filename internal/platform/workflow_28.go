package platform

import (
	"context"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
)

func (s *Service) versionSafeReport(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE platform_reports SET summary='旧版本修订',version=version+1 WHERE platform_id=? AND report_year=?`, platformID, int(time.Now().Year())); err != nil {
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
