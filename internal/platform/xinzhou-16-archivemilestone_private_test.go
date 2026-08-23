package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestArchiveMilestoneWithAudit(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestArchiveMilestoneWithAudit?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestArchiveMilestoneWithAudit@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "ArchiveMilestoneWithAudit平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestArchiveMilestoneWithAudit-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='operating',version=2 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO platform_milestones(platform_id,title,due_at,status,version) VALUES(?,?,?,'active',1)`, p.ID, "已有节点", time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	err = s.ArchiveMilestoneWithAudit(context.Background(), 1, p.ID, 2)
	if err == nil {
		t.Fatal("expected operation failure")
	}
	var got string
	if err := db.QueryRow(`SELECT status FROM innovation_platforms WHERE id=?`, p.ID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "operating" {
		t.Fatalf("platform status changed after failure: %s", got)
	}
	var rows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM platform_milestones WHERE platform_id=?`, p.ID).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("platform_milestones row leaked after failure: %d", rows)
	}
}
