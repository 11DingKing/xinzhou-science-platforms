package platform

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func TestPlanYearChangeWithAudit(t *testing.T) {
	db, err := storage.Open(context.Background(), "file:TestPlanYearChangeWithAudit?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead-TestPlanYearChangeWithAudit@example.test','x','platform_lead',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "PlanYearChangeWithAudit平台", "忻州", "先进技术", time.Now().Year()+1, 1000, "TestPlanYearChangeWithAudit-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE innovation_platforms SET status='draft',version=2 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	err = s.PlanYearChangeWithAudit(context.Background(), 1, p.ID, 2)
	if err == nil {
		t.Fatal("expected operation failure")
	}
	var got string
	if err := db.QueryRow(`SELECT status FROM innovation_platforms WHERE id=?`, p.ID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "draft" {
		t.Fatalf("platform status changed after failure: %s", got)
	}
}
