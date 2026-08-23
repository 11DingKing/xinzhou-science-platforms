package platform

import (
	"context"
	"time"
)

func (s *Service) cancelAwareMember(ctx context.Context, actorID, platformID, version int64) error {
	_, err := s.db.Exec(`INSERT INTO platform_members(platform_id,user_id,member_role,joined_at) VALUES(?,?, 'member',?)`, platformID, actorID, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
