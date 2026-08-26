package platform

import (
	"context"
	"database/sql"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func (s *Service) fundWithAudit(ctx context.Context, actorID, platformID, version int64) error {
	// The funding detail and its audit trail must commit together. If the audit
	// step fails (an approval error), the funding row is rolled back so the
	// budget balance is never consumed outside a committed approval batch.
	return storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		stamp := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `INSERT INTO platform_funding(platform_id,amount_cents,tranche,status,idempotency_key,approved_by,created_at) VALUES(?,?,?,'approved',?,?,?)`, platformID, 10, int(version), "audit-funding", actorID, stamp); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID+999, "innovation_platform", platformID, "audit", "failed", "", stamp); err != nil {
			return err
		}
		return nil
	})
}
