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

// CancelAwareMilestone must honor the expected-state invariant: a cancelled
// context (network timeout) cannot leave a milestone behind, and a frontend
// retry against the now-stale version must fail with ErrConflict instead of
// creating a duplicate node.
func TestCancelAwareMilestoneTimeoutLeavesNoNodeAndRetriesConflict(t *testing.T) {
	s := NewService(testDB(t))
	p, err := s.Create(context.Background(), 1, "掉线窗口平台", "忻州", "材料", time.Now().Year()+1, 100, "cancel-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(context.Background(), 1, p.ID, p.Version); err != nil {
		t.Fatal(err)
	}
	if err := s.StartReview(context.Background(), 2, p.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Decide(context.Background(), 2, p.ID, 3, true, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := s.Activate(context.Background(), 1, p.ID, 4); err != nil {
		t.Fatal(err)
	}
	active := int64(5)

	// Simulate a client disconnect mid-request: the context is cancelled before
	// the call, so the transaction must roll back and leave no milestone behind.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.CancelAwareMilestone(ctx, 1, p.ID, active); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM platform_milestones WHERE platform_id=?`, p.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("cancelled timeout left %d milestone(s) behind", n)
	}

	// A frontend retry against the still-current version now succeeds exactly once.
	if err := s.CancelAwareMilestone(context.Background(), 1, p.ID, active); err != nil {
		t.Fatalf("first commit failed: %v", err)
	}
	// The committed version advanced, so replaying the same version conflicts
	// instead of creating a duplicate node.
	if err := s.CancelAwareMilestone(context.Background(), 1, p.ID, active); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected conflict on stale-version retry, got %v", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM platform_milestones WHERE platform_id=?`, p.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected single committed node, got %d", n)
	}
}
