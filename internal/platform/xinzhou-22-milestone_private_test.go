package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestCancelAwareMilestone(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestCancelAwareMilestone?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestCancelAwareMilestone@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "CancelAwareMilestone平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestCancelAwareMilestone-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='approved',version=2 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = s.CancelAwareMilestone(ctx, 1, p.ID, 2)
	if err == nil {
		t.Fatal("cancelled context unexpectedly wrote platform data")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM platform_milestones WHERE platform_id=?`, p.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("cancelled request left platform_milestones rows: %d", count)
	}
}
