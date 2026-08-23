package storage

import (
	"context"
	"database/sql"
)

type Queryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func Exec(ctx context.Context, q Queryer, stmt string, args ...any) (sql.Result, error) {
	return q.ExecContext(ctx, stmt, args...)
}
func Row(ctx context.Context, q Queryer, stmt string, args ...any) *sql.Row {
	return q.QueryRowContext(ctx, stmt, args...)
}
func Rows(ctx context.Context, q Queryer, stmt string, args ...any) (*sql.Rows, error) {
	return q.QueryContext(ctx, stmt, args...)
}
