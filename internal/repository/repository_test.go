package repository

import (
	"context"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/audit"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type fixture struct {
	ctx                context.Context
	repos              *Repos
	merchant, reviewer domain.User
	version            domain.ProductVersion
	batch              domain.Batch
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "repo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := New(db)
	m, err := r.CreateUser(ctx, "merchant@example.test", "pw", domain.RoleMerchant)
	if err != nil {
		t.Fatal(err)
	}
	reviewer, err := r.CreateUser(ctx, "reviewer@example.test", "pw", domain.RoleReviewer)
	if err != nil {
		t.Fatal(err)
	}
	v, err := r.CreateVersion(ctx, m, domain.ProductVersion{SKU: "SKU-1", DisplayName: "Lamp", Channel: "online"})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.PublishVersion(ctx, m, v.ID, v.Version); err != nil {
		t.Fatal(err)
	}
	v.Status = domain.VersionPublished
	v.Version = 2
	b, err := r.CreateBatch(ctx, m, domain.Batch{VersionID: v.ID, Code: "B-1", Region: "county", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	return fixture{ctx: ctx, repos: r, merchant: m, reviewer: reviewer, version: v, batch: b}
}

func TestUserAndSessionLifecycle(t *testing.T) {
	f := newFixture(t)
	token := "token-1"
	if err := f.repos.CreateSession(f.ctx, token, f.merchant.ID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	u, err := f.repos.GetSession(f.ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != f.merchant.ID || u.Role != domain.RoleMerchant {
		t.Fatalf("session user=%+v", u)
	}
	if err := f.repos.RevokeSession(f.ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repos.GetSession(f.ctx, token); !errors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("revoked session err=%v", err)
	}
}
func TestExpiredSessionRejected(t *testing.T) {
	f := newFixture(t)
	if err := f.repos.CreateSession(f.ctx, "expired", f.merchant.ID, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repos.GetSession(f.ctx, "expired"); !errors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("expired err=%v", err)
	}
}
func TestVersionCannotPublishTwice(t *testing.T) {
	f := newFixture(t)
	if err := f.repos.PublishVersion(f.ctx, f.merchant, f.version.ID, f.version.Version); !errors.Is(err, apperrors.ErrInvalidState) {
		t.Fatalf("second publish err=%v", err)
	}
}
func TestBatchRequiresPublishedVersion(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "batch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db)
	m, _ := r.CreateUser(ctx, "m", "p", domain.RoleMerchant)
	v, _ := r.CreateVersion(ctx, m, domain.ProductVersion{SKU: "x", DisplayName: "x", Channel: "x"})
	if _, err := r.CreateBatch(ctx, m, domain.Batch{VersionID: v.ID, Code: "x", Region: "x", ExpiresAt: time.Now().Add(time.Hour)}); !errors.Is(err, apperrors.ErrInvalidState) {
		t.Fatalf("draft batch err=%v", err)
	}
}
func TestBatchStatusAndOptimisticVersion(t *testing.T) {
	f := newFixture(t)
	if err := f.repos.UpdateBatchStatus(f.ctx, f.merchant, f.batch.ID, f.batch.Version, domain.BatchSampling); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.UpdateBatchStatus(f.ctx, f.merchant, f.batch.ID, f.batch.Version, domain.BatchCleared); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("stale version err=%v", err)
	}
	if err := f.repos.UpdateBatchStatus(f.ctx, f.merchant, f.batch.ID, f.batch.Version+1, domain.BatchCleared); err != nil {
		t.Fatal(err)
	}
}
func TestInspectionClaimSubmit(t *testing.T) {
	f := newFixture(t)
	id, err := f.repos.EnsureInspection(f.ctx, f.batch.ID, f.reviewer.ID)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.repos.ClaimInspection(f.ctx, f.reviewer.ID, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id || got.Status != string(domain.InspectionLeased) {
		t.Fatalf("claim=%+v", got)
	}
	if err := f.repos.SubmitInspection(f.ctx, f.reviewer.ID, id, got.Version, "pass", "all dimensions match"); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.SubmitInspection(f.ctx, f.reviewer.ID, id, got.Version, "pass", "duplicate"); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("duplicate submit err=%v", err)
	}
}
func TestInspectionLeaseRequeue(t *testing.T) {
	f := newFixture(t)
	_, err := f.repos.EnsureInspection(f.ctx, f.batch.ID, f.reviewer.ID)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.repos.ClaimInspection(f.ctx, f.reviewer.ID, -time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.repos.RequeueExpiredInspections(f.ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	next, err := f.repos.ClaimInspection(f.ctx, f.reviewer.ID, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != got.ID {
		t.Fatalf("reclaimed %d want %d", next.ID, got.ID)
	}
}
func TestConcurrentInspectionClaimSingleWinner(t *testing.T) {
	f := newFixture(t)
	if _, err := f.repos.EnsureInspection(f.ctx, f.batch.ID, f.reviewer.ID); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := f.repos.ClaimInspection(f.ctx, f.reviewer.ID, time.Minute); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("claim winners=%d", wins)
	}
}
func TestComplaintIdempotency(t *testing.T) {
	f := newFixture(t)
	a, err := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "handle is rough", "same-request")
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "different body", "same-request")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID {
		t.Fatalf("idempotent ids %d,%d", a.ID, b.ID)
	}
}

func TestRegionalQualityUsesAuditWindow(t *testing.T) {
	f := newFixture(t)
	if _, err := f.repos.EnsureInspection(f.ctx, f.batch.ID, f.reviewer.ID); err != nil {
		t.Fatal(err)
	}
	claimed, err := f.repos.ClaimInspection(f.ctx, f.reviewer.ID, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.SubmitInspection(f.ctx, f.reviewer.ID, claimed.ID, claimed.Version, "fail", "finish differs"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "different handle", "regional-1"); err != nil {
		t.Fatal(err)
	}
	rows, err := f.repos.RegionalQuality(f.ctx, f.version.ID, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Region != "county" || rows[0].FailedInspections != 1 || rows[0].ComplaintCount != 1 || rows[0].OpenCases != 1 {
		t.Fatalf("regional quality=%+v", rows)
	}
	if !rows[0].RequiresInvestigation(1) {
		t.Fatalf("quality should require investigation: %+v", rows[0])
	}
}
func TestComplaintTransitionsAndEvidenceGate(t *testing.T) {
	f := newFixture(t)
	c, err := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "wrong finish", "c-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionComplaint(f.ctx, f.reviewer.ID, c.ID, c.Version, domain.CaseInvestigating); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionComplaint(f.ctx, f.reviewer.ID, c.ID, c.Version, domain.CaseResolved); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("stale complaint err=%v", err)
	}
	if err := f.repos.TransitionComplaint(f.ctx, f.reviewer.ID, c.ID, c.Version+1, domain.CaseResolved); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.AttachEvidence(f.ctx, f.reviewer.ID, c.ID, "photo/a", "hash"); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.CloseComplaintIfEvidenceArchived(f.ctx, f.reviewer.ID, c.ID, c.Version+2); !errors.Is(err, apperrors.ErrInvalidState) {
		t.Fatalf("closed with pending evidence err=%v", err)
	}
}
func TestComplaintCloseAfterArchive(t *testing.T) {
	f := newFixture(t)
	c, err := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "color mismatch", "c-2")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionComplaint(f.ctx, f.reviewer.ID, c.ID, c.Version, domain.CaseInvestigating); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionComplaint(f.ctx, f.reviewer.ID, c.ID, c.Version+1, domain.CaseResolved); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.AttachEvidence(f.ctx, f.reviewer.ID, c.ID, "photo/b", "hash2"); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.ArchiveEvidence(f.ctx, f.reviewer.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.CloseComplaintIfEvidenceArchived(f.ctx, f.reviewer.ID, c.ID, c.Version+2); err != nil {
		t.Fatal(err)
	}
}
func TestRemediationLifecycle(t *testing.T) {
	f := newFixture(t)
	c, err := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "batch differs", "c-3")
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.repos.CreateRemediation(f.ctx, f.reviewer.ID, c.ID, "replace supplier lot", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionRemediation(f.ctx, f.reviewer.ID, r.ID, r.Version, domain.RemediationActive); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionRemediation(f.ctx, f.reviewer.ID, r.ID, r.Version+1, domain.RemediationDone); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionRemediation(f.ctx, f.reviewer.ID, r.ID, r.Version+2, domain.RemediationActive); !errors.Is(err, apperrors.ErrInvalidState) {
		t.Fatalf("reopened err=%v", err)
	}
}
func TestExpiredRemediationEscalates(t *testing.T) {
	f := newFixture(t)
	c, _ := f.repos.OpenComplaint(f.ctx, f.merchant.ID, f.version.ID, f.batch.ID, "county", "late", "c-4")
	r, err := f.repos.CreateRemediation(f.ctx, f.reviewer.ID, c.ID, "audit", time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.TransitionRemediation(f.ctx, f.reviewer.ID, r.ID, r.Version, domain.RemediationActive); err != nil {
		t.Fatal(err)
	}
	n, err := f.repos.EscalateRemediation(f.ctx, time.Now())
	if err != nil || n != 1 {
		t.Fatalf("escalated=%d err=%v", n, err)
	}
}
func TestFulfillmentRulePublish(t *testing.T) {
	f := newFixture(t)
	from := time.Now()
	id, err := f.repos.CreateFulfillmentRule(f.ctx, f.merchant.ID, f.version.ID, "county", from, from.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.PublishFulfillmentRule(f.ctx, f.merchant.ID, id, 1); err != nil {
		t.Fatal(err)
	}
	if err := f.repos.PublishFulfillmentRule(f.ctx, f.merchant.ID, id, 1); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("republish err=%v", err)
	}
}
func TestJobsLeaseCompleteAndRetry(t *testing.T) {
	f := newFixture(t)
	if err := f.repos.Enqueue(f.ctx, "sample", "{}"); err != nil {
		t.Fatal(err)
	}
	id, err := f.repos.ClaimJob(f.ctx, "worker", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.CompleteJob(f.ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repos.ClaimJob(f.ctx, "worker", time.Minute); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("completed job claimed err=%v", err)
	}
	if err := f.repos.Enqueue(f.ctx, "retry", "{}"); err != nil {
		t.Fatal(err)
	}
	id, err = f.repos.ClaimJob(f.ctx, "worker", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repos.FailJob(f.ctx, id, "temporary"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repos.ClaimJob(f.ctx, "worker", time.Minute); err != nil {
		t.Fatalf("retry not queued: %v", err)
	}
}
func TestNotificationDelivery(t *testing.T) {
	f := newFixture(t)
	if err := f.repos.EnqueueNotification(f.ctx, f.reviewer.ID, "complaint", "{}", time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	n, err := f.repos.DeliverNotifications(f.ctx, 10, time.Now())
	if err != nil || n != 1 {
		t.Fatalf("delivered=%d err=%v", n, err)
	}
	n, err = f.repos.DeliverNotifications(f.ctx, 10, time.Now())
	if err != nil || n != 0 {
		t.Fatalf("redelivered=%d err=%v", n, err)
	}
}
func TestAuditEventsRecorded(t *testing.T) {
	f := newFixture(t)
	events, err := (audit.SQLStore{DB: f.repos.DB.DB}).List(f.ctx, "product_version", f.version.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 2 {
		t.Fatalf("audit events=%d", len(events))
	}
}
