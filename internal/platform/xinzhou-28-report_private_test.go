package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestVersionSafeReport(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestVersionSafeReport?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestVersionSafeReport@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "VersionSafeReport平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestVersionSafeReport-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='under_review',version=5 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO platform_reports(platform_id,report_year,summary,status,submitted_by,submitted_at,version) VALUES(?,?,?,'submitted',?,?,1)`, p.ID, time.Now().Year(), "原始摘要", 1, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	err = s.VersionSafeReport(context.Background(), 1, p.ID, 1)
	if err == nil {
		t.Fatal("expected stale version conflict")
	}
	var summary string
	if err := db.QueryRow(`SELECT summary FROM platform_reports WHERE platform_id=?`, p.ID).Scan(&summary); err != nil {
		t.Fatal(err)
	}
	if summary != "原始摘要" {
		t.Fatalf("stale version changed report: %s", summary)
	}
}
