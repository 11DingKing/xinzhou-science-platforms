package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type DB struct{ *sql.DB }

func Open(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	if _, err = db.ExecContext(ctx, `PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;`); err != nil {
		db.Close()
		return nil, err
	}
	if err = migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return &DB{db}, nil
}
func migrate(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	for _, e := range entries {
		b, err := migrationFiles.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		for _, stmt := range strings.Split(string(b), ";") {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("migration %s: %w", e.Name(), err)
			}
		}
	}
	return nil
}
func WithTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
