package platform

import (
	"context"
	"time"
)

func (s *Service) cancelAwareReport(ctx context.Context, actorID, platformID, version int64) error {
	_, err := s.db.Exec(`INSERT INTO platform_reports(platform_id,report_year,summary,status,submitted_by,submitted_at,version) VALUES(?,?,?,'submitted',?,?,1)`, platformID, int(time.Now().Year()), "取消后的报告", actorID, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
