package platform

import (
	"context"
	"time"
)

func (s *Service) reportWithAudit(ctx context.Context, actorID, platformID, version int64) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO platform_reports(platform_id,report_year,summary,status,submitted_by,submitted_at,version) VALUES(?,?,?,'submitted',?,?,1)`, platformID, int(time.Now().Year()), "审计报告", actorID, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID+999, "innovation_platform", platformID, "audit", "failed", "", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
