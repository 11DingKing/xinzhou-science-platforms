package platform

import (
	"context"
	"time"
)

func (s *Service) fundWithAudit(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO platform_funding(platform_id,amount_cents,tranche,status,idempotency_key,approved_by,created_at) VALUES(?,?,?,'approved',?,?,?)`, platformID, 10, int(version), "audit-funding", actorID, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID+999, "innovation_platform", platformID, "audit", "failed", "", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
