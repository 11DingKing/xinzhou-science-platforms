package platform

import (
	"context"
	"time"
)

func (s *Service) memberWithAudit(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO platform_members(platform_id,user_id,member_role,joined_at) VALUES(?,?, 'member',?)`, platformID, actorID, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID+999, "innovation_platform", platformID, "audit", "failed", "", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
