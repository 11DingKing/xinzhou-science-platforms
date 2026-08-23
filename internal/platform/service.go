package platform

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

type Status string

const (
	StatusDraft       Status = "draft"
	StatusSubmitted   Status = "submitted"
	StatusUnderReview Status = "under_review"
	StatusApproved    Status = "approved"
	StatusRejected    Status = "rejected"
	StatusOperating   Status = "operating"
	StatusCompleted   Status = "completed"
)

type MilestoneStatus string

const (
	MilestonePlanned  MilestoneStatus = "planned"
	MilestoneActive   MilestoneStatus = "active"
	MilestoneComplete MilestoneStatus = "complete"
)

type Platform struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	FocusArea   string `json:"focus_area"`
	PlannedYear int    `json:"planned_year"`
	BudgetCents int64  `json:"budget_cents"`
	Status      Status `json:"status"`
	Version     int64  `json:"version"`
}

type Milestone struct {
	ID         int64           `json:"id"`
	PlatformID int64           `json:"platform_id"`
	Title      string          `json:"title"`
	DueAt      time.Time       `json:"due_at"`
	Status     MilestoneStatus `json:"status"`
	Version    int64           `json:"version"`
}

type Service struct{ db *storage.DB }

func NewService(db *storage.DB) *Service { return &Service{db: db} }

func (s *Service) Create(ctx context.Context, actorID int64, name, city, focus string, year int, budgetCents int64, idem string) (Platform, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(city) == "" || strings.TrimSpace(focus) == "" || year < time.Now().Year() || budgetCents < 0 {
		return Platform{}, fmt.Errorf("%w: invalid platform details", apperrors.ErrInvalidState)
	}
	var out Platform
	err := storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		if idem != "" {
			var existingID int64
			err := tx.QueryRowContext(ctx, `SELECT id FROM innovation_platforms WHERE name=?`, name).Scan(&existingID)
			if err == nil {
				out, err = s.getTx(ctx, tx, existingID)
				return err
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		stamp := time.Now().UTC().Format(time.RFC3339Nano)
		res, err := tx.ExecContext(ctx, `INSERT INTO innovation_platforms(name,city,focus_area,planned_year,budget_cents,status,version,created_by,created_at,updated_at) VALUES(?,?,?,?,?,'draft',1,?,?,?)`, name, city, focus, year, budgetCents, actorID, stamp, stamp)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO platform_members(platform_id,user_id,member_role,joined_at) VALUES(?,?, 'lead',?)`, id, actorID, stamp); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID, "innovation_platform", id, "create", "ok", idem, stamp); err != nil {
			return err
		}
		out = Platform{ID: id, Name: name, City: city, FocusArea: focus, PlannedYear: year, BudgetCents: budgetCents, Status: StatusDraft, Version: 1}
		return nil
	})
	return out, err
}

func (s *Service) Submit(ctx context.Context, actorID, id, version int64) error {
	return s.transition(ctx, actorID, id, version, StatusSubmitted, StatusDraft)
}

func (s *Service) StartReview(ctx context.Context, reviewerID, id, version int64) error {
	return s.transition(ctx, reviewerID, id, version, StatusUnderReview, StatusSubmitted)
}

func (s *Service) Decide(ctx context.Context, reviewerID, id, version int64, approve bool, notes string) error {
	return storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM innovation_platforms WHERE id=?`, id).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.ErrNotFound
			}
			return err
		}
		if status != string(StatusUnderReview) {
			return apperrors.ErrInvalidState
		}
		decision := "rejected"
		next := StatusRejected
		if approve {
			decision, next = "approved", StatusApproved
		}
		stamp := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `INSERT INTO platform_reviews(platform_id,reviewer_id,decision,notes,created_at) VALUES(?,?,?,?,?)`, id, reviewerID, decision, notes, stamp); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `UPDATE innovation_platforms SET status=?,version=version+1,updated_at=? WHERE id=? AND version=?`, next, stamp, id, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, reviewerID, "innovation_platform", id, "review", decision, "", stamp)
		return err
	})
}

func (s *Service) Activate(ctx context.Context, actorID, id, version int64) error {
	return s.transition(ctx, actorID, id, version, StatusOperating, StatusApproved)
}

func (s *Service) Complete(ctx context.Context, actorID, id, version int64) error {
	return s.transition(ctx, actorID, id, version, StatusCompleted, StatusOperating)
}

func (s *Service) AddMilestone(ctx context.Context, actorID, platformID int64, title string, dueAt time.Time) (Milestone, error) {
	if strings.TrimSpace(title) == "" || dueAt.IsZero() {
		return Milestone{}, apperrors.ErrInvalidState
	}
	var out Milestone
	err := storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM innovation_platforms WHERE id=?`, platformID).Scan(&status); err != nil {
			return err
		}
		if status != string(StatusApproved) && status != string(StatusOperating) {
			return apperrors.ErrInvalidState
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO platform_milestones(platform_id,title,due_at,status,version) VALUES(?,?,?,'planned',1)`, platformID, title, dueAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		out = Milestone{ID: id, PlatformID: platformID, Title: title, DueAt: dueAt.UTC(), Status: MilestonePlanned, Version: 1}
		return nil
	})
	return out, err
}

func (s *Service) CompleteMilestone(ctx context.Context, actorID, id, version int64) error {
	res, err := s.db.ExecContext(ctx, `UPDATE platform_milestones SET status='complete',version=version+1,completed_at=? WHERE id=? AND status='active' AND version=?`, time.Now().UTC().Format(time.RFC3339Nano), id, version)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return apperrors.ErrConflict
	}
	return nil
}

func (s *Service) RecordFunding(ctx context.Context, actorID, platformID, amountCents int64, tranche int, idem string) error {
	if amountCents <= 0 || tranche < 1 || idem == "" {
		return apperrors.ErrInvalidState
	}
	return storage.WithTx(ctx, s.db.DB, func(tx *sql.Tx) error {
		var status string
		var budget, used int64
		if err := tx.QueryRowContext(ctx, `SELECT status,budget_cents FROM innovation_platforms WHERE id=?`, platformID).Scan(&status, &budget); err != nil {
			return err
		}
		if status != string(StatusApproved) && status != string(StatusOperating) {
			return apperrors.ErrInvalidState
		}
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_cents),0) FROM platform_funding WHERE platform_id=? AND status='approved'`, platformID).Scan(&used); err != nil {
			return err
		}
		if used+amountCents > budget {
			return apperrors.ErrConflict
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO platform_funding(platform_id,amount_cents,tranche,status,idempotency_key,approved_by,created_at) VALUES(?,?,?,'approved',?,?,?)`, platformID, amountCents, tranche, idem, actorID, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) SubmitReport(ctx context.Context, actorID, platformID int64, year int, summary string) error {
	if year < 2000 || strings.TrimSpace(summary) == "" {
		return apperrors.ErrInvalidState
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO platform_reports(platform_id,report_year,summary,status,submitted_by,submitted_at,version) VALUES(?,?,?,'submitted',?,?,1)`, platformID, year, summary, actorID, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Service) transition(ctx context.Context, actorID, id, version int64, next, expected Status) error {
	res, err := s.db.ExecContext(ctx, `UPDATE innovation_platforms SET status=?,version=version+1,updated_at=? WHERE id=? AND status=? AND version=?`, next, time.Now().UTC().Format(time.RFC3339Nano), id, expected, version)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return apperrors.ErrConflict
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID, "innovation_platform", id, "transition", string(next), "", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Service) getTx(ctx context.Context, tx *sql.Tx, id int64) (Platform, error) {
	var p Platform
	err := tx.QueryRowContext(ctx, `SELECT id,name,city,focus_area,planned_year,budget_cents,status,version FROM innovation_platforms WHERE id=?`, id).Scan(&p.ID, &p.Name, &p.City, &p.FocusArea, &p.PlannedYear, &p.BudgetCents, &p.Status, &p.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return p, apperrors.ErrNotFound
	}
	return p, err
}

func (s *Service) CancelAwareFunding(ctx context.Context, actorID, platformID, version int64) error {
	return s.cancelAwareFunding(ctx, actorID, platformID, version)
}
