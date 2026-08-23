package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func openTestDB(t *testing.T) (*DB, context.Context) {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, "file:"+filepath.Join(t.TempDir(), "quality.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, ctx
}

func TestOpenCreatesAllTables(t *testing.T) {
	db, ctx := openTestDB(t)
	rows, err := db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		names = append(names, n)
	}
	want := []string{"audit_events", "batches", "complaints", "evidence", "fulfillment_rules", "idempotency_keys", "innovation_platforms", "inspections", "jobs", "notifications", "platform_funding", "platform_members", "platform_milestones", "platform_reports", "platform_reviews", "product_versions", "remediation", "sessions", "users"}
	if len(names) != len(want) {
		t.Fatalf("tables=%v want=%v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("table[%d]=%s want=%s", i, names[i], want[i])
		}
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	db, ctx := openTestDB(t)
	if err := migrate(ctx, db.DB); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("unexpected seeded rows=%d", count)
	}
}

func TestForeignKeysRejectOrphanBatch(t *testing.T) {
	db, ctx := openTestDB(t)
	_, err := db.ExecContext(ctx, `INSERT INTO batches(version_id,code,region,status,expires_at,version,created_at) VALUES(999,'b','r','pending','2026-08-24T00:00:00Z',1,'now')`)
	if err == nil {
		t.Fatal("orphan batch accepted")
	}
	if !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	db, ctx := openTestDB(t)
	err := WithTx(ctx, db.DB, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO users(email,password_hash,role,created_at) VALUES('a','p','merchant','now')`); err != nil {
			return err
		}
		return errors.New("forced failure")
	})
	if err == nil {
		t.Fatal("transaction error lost")
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback leaked %d rows", count)
	}
}

func TestWithTxCommitsMultipleRows(t *testing.T) {
	db, ctx := openTestDB(t)
	if err := WithTx(ctx, db.DB, func(tx *sql.Tx) error {
		for _, email := range []string{"one", "two"} {
			if _, err := tx.ExecContext(ctx, `INSERT INTO users(email,password_hash,role,created_at) VALUES(?,?,?,?)`, email, "p", "merchant", "now"); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("committed rows=%d", count)
	}
}
