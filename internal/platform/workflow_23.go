package platform

import (
	"context"
	"time"
)

func (s *Service) cancelAwareFunding(ctx context.Context, actorID, platformID, version int64) error {
	_, err := s.db.Exec(`INSERT INTO platform_funding(platform_id,amount_cents,tranche,status,idempotency_key,approved_by,created_at) VALUES(?,?,?,'approved',?,?,?)`, platformID, 10, int(version), "approved", actorID, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
