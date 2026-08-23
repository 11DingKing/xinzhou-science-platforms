package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestVersionSafeMilestone(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestVersionSafeMilestone?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestVersionSafeMilestone@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "VersionSafeMilestone平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestVersionSafeMilestone-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='under_review',version=5 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO platform_milestones(platform_id,title,due_at,status,version) VALUES(?,?,?,'active',1)`, p.ID, "旧节点", time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	err = s.VersionSafeMilestone(context.Background(), 1, p.ID, 1)
	if err == nil {
		t.Fatal("expected stale version conflict")
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM platform_milestones WHERE platform_id=?`, p.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "active" {
		t.Fatalf("stale version changed milestone: %s", status)
	}
}
