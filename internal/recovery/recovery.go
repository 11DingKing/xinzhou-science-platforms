package recovery

import (
	"context"
	"database/sql"
	"time"
)

type Summary struct{ LeasesRequeued, JobsRequeued, RemediationsEscalated int }
type Clock func() time.Time

func Scan(ctx context.Context, db *sql.DB, now time.Time) (Summary, error) {
	var out Summary
	res, err := db.ExecContext(ctx, `UPDATE inspections SET status='queued',reviewer_id=NULL,lease_until=NULL,version=version+1 WHERE status='leased' AND lease_until<=?`, now.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return out, err
	}
	n, _ := res.RowsAffected()
	out.LeasesRequeued = int(n)
	res, err = db.ExecContext(ctx, `UPDATE jobs SET status='queued',lease_until=NULL,updated_at=? WHERE status='leased' AND lease_until<=?`, now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return out, err
	}
	n, _ = res.RowsAffected()
	out.JobsRequeued = int(n)
	res, err = db.ExecContext(ctx, `UPDATE remediation SET status='escalated',version=version+1 WHERE status='active' AND due_at<?`, now.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return out, err
	}
	n, _ = res.RowsAffected()
	out.RemediationsEscalated = int(n)
	return out, nil
}
func Start(ctx context.Context, db *sql.DB, interval time.Duration, clock Clock, done chan<- Summary) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			summary, err := Scan(ctx, db, clock())
			if err == nil && done != nil {
				done <- summary
			}
		}
	}
}
