package transaction

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"testing"
)

func TestTransactionRunner(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "tx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := Runner{DB: db.DB}
	if err := r.Do(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO users(email,password_hash,role,created_at) VALUES('tx','p','merchant','now')`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.Do(ctx, func(context.Context, *sql.Tx) error { return errors.New("abort") }); err == nil {
		t.Fatal("abort")
	}
}
