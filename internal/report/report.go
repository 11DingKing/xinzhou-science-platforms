package report

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

type BatchRow struct {
	ID                                 int64
	Code, Region, Status, SKU, Channel string
	ComplaintCount, InspectionCount    int
	CreatedAt                          time.Time
}
type ComplaintRow struct {
	ID                          int64
	Region, Status, Description string
	SKU                         string
	BatchCode                   string
	CreatedAt                   time.Time
}
type Filter struct {
	Search, Region, Status string
	From, To               time.Time
	Limit, Offset          int
}

func (f Filter) Normalize() Filter {
	out := f
	if out.Limit < 1 {
		out.Limit = 20
	}
	if out.Limit > 100 {
		out.Limit = 100
	}
	if out.Offset < 0 {
		out.Offset = 0
	}
	out.Search = strings.TrimSpace(out.Search)
	out.Region = strings.TrimSpace(out.Region)
	out.Status = strings.TrimSpace(out.Status)
	return out
}
func BuildBatchQuery(f Filter) (string, []any) {
	f = f.Normalize()
	clauses := []string{"1=1"}
	args := []any{}
	if f.Search != "" {
		clauses = append(clauses, "(b.code LIKE ? OR p.sku LIKE ? OR p.display_name LIKE ?)")
		q := "%" + f.Search + "%"
		args = append(args, q, q, q)
	}
	if f.Region != "" {
		clauses = append(clauses, "b.region=?")
		args = append(args, f.Region)
	}
	if f.Status != "" {
		clauses = append(clauses, "b.status=?")
		args = append(args, f.Status)
	}
	if !f.From.IsZero() {
		clauses = append(clauses, "b.created_at>=?")
		args = append(args, f.From.UTC().Format(time.RFC3339Nano))
	}
	if !f.To.IsZero() {
		clauses = append(clauses, "b.created_at<?")
		args = append(args, f.To.UTC().Format(time.RFC3339Nano))
	}
	q := `SELECT b.id,b.code,b.region,b.status,p.sku,p.channel,b.created_at,COUNT(DISTINCT c.id),COUNT(DISTINCT i.id) FROM batches b JOIN product_versions p ON p.id=b.version_id LEFT JOIN complaints c ON c.batch_id=b.id LEFT JOIN inspections i ON i.batch_id=b.id WHERE ` + strings.Join(clauses, " AND ") + ` GROUP BY b.id ORDER BY b.created_at DESC,b.id DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)
	return q, args
}
func BuildComplaintQuery(f Filter) (string, []any) {
	f = f.Normalize()
	clauses := []string{"1=1"}
	args := []any{}
	if f.Search != "" {
		clauses = append(clauses, "(c.description LIKE ? OR c.region LIKE ? OR p.sku LIKE ?)")
		q := "%" + f.Search + "%"
		args = append(args, q, q, q)
	}
	if f.Region != "" {
		clauses = append(clauses, "c.region=?")
		args = append(args, f.Region)
	}
	if f.Status != "" {
		clauses = append(clauses, "c.status=?")
		args = append(args, f.Status)
	}
	q := `SELECT c.id,c.region,c.status,c.description,p.sku,b.code,c.created_at FROM complaints c JOIN product_versions p ON p.id=c.version_id JOIN batches b ON b.id=c.batch_id WHERE ` + strings.Join(clauses, " AND ") + ` ORDER BY c.created_at DESC,c.id DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)
	return q, args
}
func ListBatches(ctx context.Context, db *sql.DB, f Filter) ([]BatchRow, error) {
	q, args := BuildBatchQuery(f)
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BatchRow{}
	for rows.Next() {
		var row BatchRow
		var ts string
		if err := rows.Scan(&row.ID, &row.Code, &row.Region, &row.Status, &row.SKU, &row.Channel, &ts, &row.ComplaintCount, &row.InspectionCount); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, row)
	}
	return out, rows.Err()
}
func ListComplaints(ctx context.Context, db *sql.DB, f Filter) ([]ComplaintRow, error) {
	q, args := BuildComplaintQuery(f)
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ComplaintRow{}
	for rows.Next() {
		var row ComplaintRow
		var ts string
		if err := rows.Scan(&row.ID, &row.Region, &row.Status, &row.Description, &row.SKU, &row.BatchCode, &ts); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, row)
	}
	return out, rows.Err()
}
func SummarizeRegions(rows []BatchRow) map[string]int {
	out := map[string]int{}
	for _, row := range rows {
		out[row.Region]++
	}
	return out
}
func StatusCounts(rows []ComplaintRow) []string {
	out := []string{}
	for _, row := range rows {
		out = append(out, row.Status)
	}
	sort.Strings(out)
	return out
}
func ValidateFilter(f Filter) error {
	if f.Limit > 100 {
		return fmt.Errorf("limit exceeds maximum")
	}
	if !f.From.IsZero() && !f.To.IsZero() && !f.To.After(f.From) {
		return fmt.Errorf("invalid time range")
	}
	return nil
}
