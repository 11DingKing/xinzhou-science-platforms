package platform

import (
	"context"
	"database/sql"
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

// approvedFundingCents sums the platform_funding rows that still count toward
// the used budget, mirroring RecordFunding's quota computation.
func approvedFundingCents(t *testing.T, db *storage.DB, platformID int64) int64 {
	t.Helper()
	var used int64
	if err := db.QueryRow(`SELECT COALESCE(SUM(amount_cents),0) FROM platform_funding WHERE platform_id=? AND status='approved'`, platformID).Scan(&used); err != nil {
		t.Fatal(err)
	}
	return used
}

// auditCancelFundingResult returns the result of the audit event recorded for a
// cancel_funding action on the platform, if any.
func auditCancelFundingResult(t *testing.T, db *storage.DB, platformID int64) (string, bool) {
	t.Helper()
	var result string
	err := db.QueryRow(`SELECT result FROM audit_events WHERE object_id=? AND action='cancel_funding' ORDER BY id DESC LIMIT 1`, platformID).Scan(&result)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return result, true
}

func TestCancelFundingRevokesRealVoucherAtomically(t *testing.T) {
	db := testDB(t)
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "撤销-平台", "忻州", "材料", time.Now().Year()+1, 1000, "rev-1")
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
	if err := s.RecordFunding(context.Background(), 1, p.ID, 600, 1, "rev-fund-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Activate(context.Background(), 1, p.ID, 4); err != nil {
		t.Fatal(err)
	}

	// Revoke on the correct version succeeds and reflects the real voucher:
	// the approved funding flips to cancelled, reclaiming the budget rather
	// than deducting without a voucher.
	if err := s.CancelFundingWithAudit(context.Background(), 1, p.ID, 5); err != nil {
		t.Fatalf("expected cancel to succeed, got %v", err)
	}
	if used := approvedFundingCents(t, db, p.ID); used != 0 {
		t.Fatalf("approved funding should be reclaimed, got used=%d", used)
	}
	if res, ok := auditCancelFundingResult(t, db, p.ID); !ok || res != "ok" {
		t.Fatalf("expected ok cancel_funding audit event, got %q ok=%v", res, ok)
	}

	// After reclaim, the full budget is available again.
	if err := s.RecordFunding(context.Background(), 1, p.ID, 1000, 2, "rev-fund-2"); err != nil {
		t.Fatalf("expected budget reclaimed after cancel, got %v", err)
	}
}

func TestCancelFundingRequiresRealVoucherAndIsAtomic(t *testing.T) {
	db := testDB(t)
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "撤销-无凭证", "忻州", "材料", time.Now().Year()+1, 1000, "rev-2")
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

	// No approved funding exists yet: cancelling must refuse rather than
	// fabricate a voucher-less deduction, and must change nothing.
	beforeUsed := approvedFundingCents(t, db, p.ID)
	if !errors.Is(s.CancelFundingWithAudit(context.Background(), 1, p.ID, 4), apperrors.ErrInvalidState) {
		t.Fatal("expected invalid state when no real voucher exists")
	}
	if used := approvedFundingCents(t, db, p.ID); used != beforeUsed {
		t.Fatalf("approved funding must not change on refused cancel, before=%d after=%d", beforeUsed, used)
	}
}

func TestCancelFundingRejectsStaleVersion(t *testing.T) {
	db := testDB(t)
	s := NewService(db)
	p, err := s.Create(context.Background(), 1, "撤销-版本", "忻州", "材料", time.Now().Year()+1, 1000, "rev-3")
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
	if err := s.RecordFunding(context.Background(), 1, p.ID, 600, 1, "rev-fund-3"); err != nil {
		t.Fatal(err)
	}
	// Stale version: must return conflict and leave funding approved.
	if !errors.Is(s.CancelFundingWithAudit(context.Background(), 1, p.ID, 3), apperrors.ErrConflict) {
		t.Fatal("expected conflict on stale version")
	}
	if used := approvedFundingCents(t, db, p.ID); used != 600 {
		t.Fatalf("funding must remain approved on stale-version cancel, got used=%d", used)
	}
}
