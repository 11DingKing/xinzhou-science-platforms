package health

import (
	"context"
	"database/sql"
	"time"
)

type Status struct {
	Service   string
	Ready     bool
	Database  bool
	CheckedAt time.Time
	Message   string
}

func Check(ctx context.Context, db *sql.DB, now time.Time) Status {
	status := Status{Service: "ab-quality", CheckedAt: now.UTC()}
	if db == nil {
		status.Message = "database unavailable"
		return status
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		status.Message = err.Error()
		return status
	}
	status.Ready = true
	status.Database = true
	status.Message = "ready"
	return status
}
func Alive(now time.Time) Status {
	return Status{Service: "ab-quality", Ready: true, Database: false, CheckedAt: now.UTC(), Message: "alive"}
}
func Degraded(message string, now time.Time) Status {
	return Status{Service: "ab-quality", Ready: false, Database: false, CheckedAt: now.UTC(), Message: message}
}
