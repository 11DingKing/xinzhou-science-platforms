package audit

import (
	"context"
	"database/sql"
	"time"
)

type RetentionReport struct{ Deleted int }

func Purge(ctx context.Context, db *sql.DB, before time.Time) (RetentionReport, error) {
	res, err := db.ExecContext(ctx, `DELETE FROM audit_events WHERE created_at<?`, before.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return RetentionReport{}, err
	}
	n, _ := res.RowsAffected()
	return RetentionReport{Deleted: int(n)}, nil
}
func Count(ctx context.Context, db *sql.DB, typ string, id int64) (int, error) {
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_events WHERE object_type=? AND object_id=?`, typ, id).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
