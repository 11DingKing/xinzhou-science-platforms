package audit

import (
	"context"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func TestAuditRetentionEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if n, err := Count(ctx, db.DB, "x", 1); err != nil || n != 0 {
		t.Fatalf("count=%d err=%v", n, err)
	}
	if out, err := Purge(ctx, db.DB, time.Now()); err != nil || out.Deleted != 0 {
		t.Fatalf("purge=%+v err=%v", out, err)
	}
}
