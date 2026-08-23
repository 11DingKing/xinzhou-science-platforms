package platform

import (
	"context"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
)

func (s *Service) versionSafeFunding(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO platform_funding(platform_id,amount_cents,tranche,status,idempotency_key,approved_by,created_at) VALUES(?,?,?,'approved',?,?,?)`, platformID, 10, int(version), "version-guard", actorID, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
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
