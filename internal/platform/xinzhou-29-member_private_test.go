package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestVersionSafeMember(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestVersionSafeMember?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestVersionSafeMember@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "VersionSafeMember平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestVersionSafeMember-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='under_review',version=5 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	err = s.VersionSafeMember(context.Background(), 1, p.ID, 1)
	if err == nil {
		t.Fatal("expected stale version conflict")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM platform_members WHERE platform_id=?`, p.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("stale version left platform_members rows: %d", count)
	}
}
