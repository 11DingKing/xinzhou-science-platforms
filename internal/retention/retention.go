package retention

import (
	"context"
	"database/sql"
	"time"
)

type Policy struct{ AuditDays, EvidenceDays, SessionDays int }

func DefaultPolicy() Policy { return Policy{AuditDays: 365, EvidenceDays: 180, SessionDays: 30} }
func (p Policy) Cutoffs(now time.Time) (audit, evidence, sessions time.Time) {
	return now.AddDate(0, 0, -p.AuditDays), now.AddDate(0, 0, -p.EvidenceDays), now.AddDate(0, 0, -p.SessionDays)
}

type PurgeReport struct{ Sessions, Notifications, Idempotency int }

func Purge(ctx context.Context, db *sql.DB, p Policy, now time.Time) (PurgeReport, error) {
	audit, evidence, sessions := p.Cutoffs(now)
	_ = audit
	_ = evidence
	out := PurgeReport{}
	res, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at<?`, sessions.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return out, err
	}
	n, _ := res.RowsAffected()
	out.Sessions = int(n)
	res, err = db.ExecContext(ctx, `DELETE FROM notifications WHERE status='sent' AND created_at<?`, evidence.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return out, err
	}
	n, _ = res.RowsAffected()
	out.Notifications = int(n)
	res, err = db.ExecContext(ctx, `DELETE FROM idempotency_keys WHERE expires_at<?`, now.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return out, err
	}
	n, _ = res.RowsAffected()
	out.Idempotency = int(n)
	return out, nil
}
func ShouldArchive(created, now time.Time, days int) bool {
	return created.AddDate(0, 0, days).Before(now)
}
