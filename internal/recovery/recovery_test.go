package recovery

import (
	"context"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func TestRecoveryRequeuesExpiredState(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "recover.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := repository.New(db)
	if err := r.Enqueue(ctx, "sample", "{}"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ClaimJob(ctx, "worker", -time.Minute); err != nil {
		t.Fatal(err)
	}
	summary, err := Scan(ctx, db.DB, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if summary.JobsRequeued != 1 {
		t.Fatalf("summary=%+v", summary)
	}
}
