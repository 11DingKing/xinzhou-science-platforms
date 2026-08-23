package orchestrator

import (
	"context"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkflowReleaseQueuesInspection(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := repository.New(db)
	m, _ := r.CreateUser(ctx, "merchant", "p", domain.RoleMerchant)
	reviewer, _ := r.CreateUser(ctx, "reviewer", "p", domain.RoleReviewer)
	v, _ := r.CreateVersion(ctx, m, domain.ProductVersion{SKU: "s", DisplayName: "name", Channel: "online"})
	if err := r.PublishVersion(ctx, m, v.ID, v.Version); err != nil {
		t.Fatal(err)
	}
	v.Status = domain.VersionPublished
	v.Version++
	b, _ := r.CreateBatch(ctx, m, domain.Batch{VersionID: v.ID, Code: "b", Region: "city", ExpiresAt: time.Now().Add(time.Hour)})
	w := New(r)
	if _, err := w.Release(ctx, m, ReleaseInput{Version: v, Batch: b, ReviewerID: reviewer.ID}); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM batches WHERE id=?`, b.ID).Scan(&status); err != nil || status != "sampling" {
		t.Fatalf("status=%s err=%v", status, err)
	}
}
