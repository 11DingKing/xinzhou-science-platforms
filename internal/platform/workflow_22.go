package platform

import (
	"context"
	"time"
)

func (s *Service) cancelAwareMilestone(ctx context.Context, actorID, platformID, version int64) error {
	_, err := s.db.Exec(`INSERT INTO platform_milestones(platform_id,title,due_at,status,version) VALUES(?,?,?,'planned',1)`, platformID, "取消后的节点", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
