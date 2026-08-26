package platform

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

// cancelFundingWithAudit revokes a platform's approved funding appropriations on
// the revocation review desk.
//
// It must reflect a real voucher: the existing approved platform_funding rows
// are flipped to "cancelled" so they no longer count toward the used budget
// (a reclaim, not a voucher-less deduction), and exactly one audit event is
// recorded for the real actor with a success result. The whole operation runs
// in one transaction, so a returned error guarantees no funding state changed
// — preserving the platform's expected-state invariants.
func (s *Service) cancelFundingWithAudit(ctx context.Context, actorID, platformID, version int64) error {
	return storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM innovation_platforms WHERE id=?`, platformID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.ErrNotFound
			}
			return err
		}
		if status != string(StatusApproved) && status != string(StatusOperating) {
			return apperrors.ErrInvalidState
		}
		stamp := time.Now().UTC().Format(time.RFC3339Nano)
		// Optimistic-concurrency guard: refuse if the platform was modified concurrently.
		res, err := tx.ExecContext(ctx, `UPDATE innovation_platforms SET version=version+1, updated_at=? WHERE id=? AND version=?`, stamp, platformID, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		// Flip the real approved funding vouchers to "cancelled"; they are then
		// excluded from the used-budget sum, reclaiming the budget rather than
		// deducting without a voucher.
		fundRes, err := tx.ExecContext(ctx, `UPDATE platform_funding SET status='cancelled' WHERE platform_id=? AND status='approved'`, platformID)
		if err != nil {
			return err
		}
		if n, _ := fundRes.RowsAffected(); n == 0 {
			// No real voucher to revoke — refuse rather than fabricate one.
			return apperrors.ErrInvalidState
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID, "innovation_platform", platformID, "cancel_funding", "ok", "", stamp)
		return err
	})
}
