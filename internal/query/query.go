package query

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Page struct {
	Number int
	Size   int
}

func (p Page) Normalize() Page {
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = 20
	}
	if p.Size > 100 {
		p.Size = 100
	}
	return p
}
func (p Page) Offset() int { p = p.Normalize(); return (p.Number - 1) * p.Size }

type Sort struct {
	Field      string
	Descending bool
}

func (s Sort) Clause(allowed map[string]string) (string, error) {
	column, ok := allowed[s.Field]
	if !ok {
		return "", errors.New("sort field not allowed")
	}
	direction := "ASC"
	if s.Descending {
		direction = "DESC"
	}
	return column + " " + direction, nil
}
func Count(ctx context.Context, db *sql.DB, table, where string, args ...any) (int, error) {
	if strings.TrimSpace(table) == "" {
		return 0, errors.New("table required")
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE "+where, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
func ScanStringRows(rows *sql.Rows) ([]string, error) {
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}
