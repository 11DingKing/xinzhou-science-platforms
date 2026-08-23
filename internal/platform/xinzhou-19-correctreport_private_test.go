package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestCorrectReportWithAudit(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestCorrectReportWithAudit?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestCorrectReportWithAudit@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "CorrectReportWithAudit平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestCorrectReportWithAudit-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='operating',version=2 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO platform_reports(platform_id,report_year,summary,status,submitted_by,submitted_at,version) VALUES(?,?,?,'submitted',?,?,1)`, p.ID, time.Now().Year(), "原始摘要", 1, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	err = s.CorrectReportWithAudit(context.Background(), 1, p.ID, 2)
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
	if err := db.QueryRow(`SELECT COUNT(*) FROM platform_reports WHERE platform_id=?`, p.ID).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("platform_reports row leaked after failure: %d", rows)
	}
}
