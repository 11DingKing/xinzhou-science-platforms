package platform

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

func testDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.Open(context.Background(), "file:platform-test-"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users(id,email,password_hash,role,created_at) VALUES(1,'lead@example.test','x','platform_lead',CURRENT_TIMESTAMP),(2,'admin@example.test','x','platform_admin',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndLifecycle(t *testing.T) {
	s := NewService(testDB(t))
	p, err := s.Create(context.Background(), 1, "忻州先进材料平台", "忻州", "先进材料", time.Now().Year()+1, 1000, "apply-1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != StatusDraft || p.Version != 1 {
		t.Fatalf("unexpected platform: %+v", p)
	}
	if err := s.Submit(context.Background(), 1, p.ID, p.Version); err != nil {
		t.Fatal(err)
	}
	if err := s.StartReview(context.Background(), 2, p.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Decide(context.Background(), 2, p.ID, 3, true, "通过"); err != nil {
		t.Fatal(err)
	}
	if err := s.Activate(context.Background(), 1, p.ID, 4); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddMilestone(context.Background(), 1, p.ID, "中试线投产", time.Now().Add(30*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordFunding(context.Background(), 1, p.ID, 500, 1, "fund-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitReport(context.Background(), 1, p.ID, time.Now().Year(), "年度建设完成"); err != nil {
		t.Fatal(err)
	}
}

func TestTransitionConflictAndValidation(t *testing.T) {
	s := NewService(testDB(t))
	if _, err := s.Create(context.Background(), 1, "", "忻州", "材料", time.Now().Year()+1, 100, "x"); !errors.Is(err, apperrors.ErrInvalidState) {
		t.Fatalf("expected validation error, got %v", err)
	}
	p, err := s.Create(context.Background(), 1, "平台", "忻州", "材料", time.Now().Year()+1, 100, "y")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(context.Background(), 1, p.ID, 99); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if err := s.StartReview(context.Background(), 2, p.ID, p.Version); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected invalid version conflict, got %v", err)
	}
}

func TestFundingQuotaAndReportUniqueness(t *testing.T) {
	s := NewService(testDB(t))
	p, err := s.Create(context.Background(), 1, "平台-资金", "忻州", "能源", time.Now().Year()+1, 100, "z")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(context.Background(), 1, p.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.StartReview(context.Background(), 2, p.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Decide(context.Background(), 2, p.ID, 3, true, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordFunding(context.Background(), 1, p.ID, 90, 1, "f-1"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(s.RecordFunding(context.Background(), 1, p.ID, 20, 2, "f-2"), apperrors.ErrConflict) {
		t.Fatal("quota should reject overspend")
	}
	if err := s.SubmitReport(context.Background(), 1, p.ID, 2026, "a"); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitReport(context.Background(), 1, p.ID, 2026, "b"); err == nil {
		t.Fatal("duplicate annual report should fail")
	}
}

// FundWithAudit's audit step references an actor that does not exist, so the
// approval fails. The funding detail must roll back together with the audit so
// the budget balance is not consumed by a failed approval batch.
func TestFundWithAuditRollsBackBudgetOnApprovalError(t *testing.T) {
	db := testDB(t)
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "平台-审批", "忻州", "材料", time.Now().Year()+1, 100, "audit-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(context.Background(), 1, p.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.StartReview(context.Background(), 2, p.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Decide(context.Background(), 2, p.ID, 3, true, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := s.FundWithAudit(context.Background(), 1, p.ID, p.Version); err == nil {
		t.Fatal("expected FundWithAudit approval to fail on the audit step")
	}
	var used int64
	if err := db.QueryRow(`SELECT COALESCE(SUM(amount_cents),0) FROM platform_funding WHERE platform_id=? AND status='approved'`, p.ID).Scan(&used); err != nil {
		t.Fatal(err)
	}
	if used != 0 {
		t.Fatalf("budget consumed by failed approval: used=%d", used)
	}
}
