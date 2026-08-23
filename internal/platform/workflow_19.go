package platform

import (
	"context"
	"time"
)

func (s *Service) correctReportWithAudit(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE platform_reports SET summary='审计失败后的摘要',version=version+1 WHERE platform_id=?`, platformID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID+999, "innovation_platform", platformID, "audit", "failed", "", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
