package platform

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

// cancelAwareMilestone records a construction node for a platform while
// honouring the expected-state invariant.
//
// The whole step runs in a single context-aware transaction, so a network
// timeout (cancelled context) cannot leave a half-applied ledger entry behind:
// if the client disconnects before the transaction commits, no milestone is
// created. Advancing the platform version on success means a frontend retry
// against the now-stale version fails the guard with ErrConflict instead of
// creating a duplicate node, so the platform keeps obeying its expected-state
// invariant.
func (s *Service) cancelAwareMilestone(ctx context.Context, actorID, platformID, version int64) error {
	return storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		var currentVersion int64
		err := tx.QueryRowContext(ctx, `SELECT version FROM innovation_platforms WHERE id=?`, platformID).Scan(&currentVersion)
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if err != nil {
			return err
		}
		if currentVersion != version {
			return apperrors.ErrConflict
		}
		stamp := time.Now().UTC().Format(time.RFC3339Nano)
		res, err := tx.ExecContext(ctx, `UPDATE innovation_platforms SET version=version+1, updated_at=? WHERE id=? AND version=?`, stamp, platformID, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO platform_milestones(platform_id,title,due_at,status,version) VALUES(?,?,?,'planned',1)`, platformID, "取消后的节点", stamp); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID, "innovation_platform", platformID, "cancel_milestone", "ok", "", stamp); err != nil {
			return err
		}
		return nil
	})
}
