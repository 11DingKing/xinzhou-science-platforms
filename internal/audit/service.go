package audit

import (
	"context"
	"database/sql"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"time"
)

type Store interface {
	Append(context.Context, domain.AuditEvent) error
	List(context.Context, string, int64) ([]domain.AuditEvent, error)
}
type Service struct {
	store Store
	clock func() time.Time
}

func New(s Store) *Service { return &Service{store: s, clock: time.Now} }
func (s *Service) Record(ctx context.Context, actor int64, typ string, objectID int64, action, result, requestID string) error {
	return s.store.Append(ctx, domain.AuditEvent{ActorID: actor, ObjectType: typ, ObjectID: objectID, Action: action, Result: result, RequestID: requestID, CreatedAt: s.clock().UTC()})
}
func (s *Service) Recent(ctx context.Context, typ string, objectID int64) ([]domain.AuditEvent, error) {
	return s.store.List(ctx, typ, objectID)
}

type SQLStore struct{ DB *sql.DB }

func (s SQLStore) Append(ctx context.Context, e domain.AuditEvent) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, e.ActorID, e.ObjectType, e.ObjectID, e.Action, e.Result, e.RequestID, e.CreatedAt.Format(time.RFC3339Nano))
	return err
}
func (s SQLStore) List(ctx context.Context, typ string, id int64) ([]domain.AuditEvent, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,actor_id,object_type,object_id,action,result,request_id,created_at FROM audit_events WHERE object_type=? AND object_id=? ORDER BY id`, typ, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		var ts string
		if err := rows.Scan(&e.ID, &e.ActorID, &e.ObjectType, &e.ObjectID, &e.Action, &e.Result, &e.RequestID, &ts); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, e)
	}
	return out, rows.Err()
}
