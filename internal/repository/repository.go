package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
)

type Repos struct{ DB *storage.DB }

func New(db *storage.DB) *Repos { return &Repos{DB: db} }
func now() string               { return time.Now().UTC().Format(time.RFC3339Nano) }
func (r *Repos) CreateUser(ctx context.Context, email, pass string, role domain.Role) (domain.User, error) {
	res, err := r.DB.ExecContext(ctx, `INSERT INTO users(email,password_hash,role,created_at) VALUES(?,?,?,?)`, email, pass, role, now())
	if err != nil {
		return domain.User{}, err
	}
	id, _ := res.LastInsertId()
	return domain.User{ID: id, Email: email, PasswordHash: pass, Role: role}, nil
}
func (r *Repos) FindUser(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	var disabled int
	err := r.DB.QueryRowContext(ctx, `SELECT id,email,password_hash,role,disabled FROM users WHERE email=?`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return u, apperrors.ErrNotFound
	}
	u.Disabled = disabled != 0
	return u, err
}
func (r *Repos) CreateSession(ctx context.Context, token string, userID int64, expires time.Time) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO sessions(id,user_id,expires_at,created_at) VALUES(?,?,?,?)`, token, userID, expires.UTC().Format(time.RFC3339Nano), now())
	return err
}
func (r *Repos) GetSession(ctx context.Context, token string) (domain.User, error) {
	var u domain.User
	var exp string
	var disabled int
	err := r.DB.QueryRowContext(ctx, `SELECT u.id,u.email,u.password_hash,u.role,u.disabled,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.id=? AND s.revoked_at IS NULL`, token).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &disabled, &exp)
	if errors.Is(err, sql.ErrNoRows) {
		return u, apperrors.ErrUnauthorized
	}
	t, err := time.Parse(time.RFC3339Nano, exp)
	if err != nil || time.Now().After(t) {
		return u, apperrors.ErrUnauthorized
	}
	u.Disabled = disabled != 0
	return u, nil
}
func (r *Repos) RevokeSession(ctx context.Context, token string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE sessions SET revoked_at=? WHERE id=?`, now(), token)
	return err
}
func (r *Repos) CreateVersion(ctx context.Context, u domain.User, v domain.ProductVersion) (domain.ProductVersion, error) {
	var out domain.ProductVersion
	err := storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `INSERT INTO product_versions(merchant_id,sku,display_name,channel,status,version,created_at) VALUES(?,?,?,?,?,?,?)`, u.ID, v.SKU, v.DisplayName, v.Channel, domain.VersionDraft, 1, now())
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		out = v
		out.ID = id
		out.MerchantID = u.ID
		out.Status = domain.VersionDraft
		out.Version = 1
		return r.auditTx(ctx, tx, u.ID, "product_version", id, "create", "ok", "")
	})
	return out, err
}
func (r *Repos) PublishVersion(ctx context.Context, actor domain.User, id, version int64) error {
	return storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM product_versions WHERE id=?`, id).Scan(&status); err != nil {
			return err
		}
		if status != string(domain.VersionDraft) {
			return apperrors.ErrInvalidState
		}
		if _, err := tx.ExecContext(ctx, `UPDATE product_versions SET status=?,version=version+1 WHERE id=? AND version=?`, domain.VersionPublished, id, version); err != nil {
			return err
		}
		return r.auditTx(ctx, tx, actor.ID, "product_version", id, "publish", "ok", "")
	})
}
func (r *Repos) CreateBatch(ctx context.Context, actor domain.User, b domain.Batch) (domain.Batch, error) {
	var out domain.Batch
	err := storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM product_versions WHERE id=?`, b.VersionID).Scan(&status); err != nil {
			return err
		}
		if status != string(domain.VersionPublished) {
			return apperrors.ErrInvalidState
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO batches(version_id,code,region,status,expires_at,version,created_at) VALUES(?,?,?,?,?,?,?)`, b.VersionID, b.Code, b.Region, domain.BatchPending, b.ExpiresAt.UTC().Format(time.RFC3339Nano), 1, now())
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		out = b
		out.ID = id
		out.Status = domain.BatchPending
		out.Version = 1
		return r.auditTx(ctx, tx, actor.ID, "batch", id, "create", "ok", "")
	})
	return out, err
}
func (r *Repos) ClaimJob(ctx context.Context, workerID string, lease time.Duration) (int64, error) {
	var id int64
	err := storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, `SELECT id FROM jobs WHERE status='queued' OR (status='leased' AND lease_until<?) ORDER BY id LIMIT 1`, now())
		if err := row.Scan(&id); err != nil {
			return apperrors.ErrNotFound
		}
		_, err := tx.ExecContext(ctx, `UPDATE jobs SET status='leased',lease_until=?,updated_at=? WHERE id=?`, time.Now().Add(lease).UTC().Format(time.RFC3339Nano), now(), id)
		return err
	})
	return id, err
}
func (r *Repos) Enqueue(ctx context.Context, kind, payload string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO jobs(kind,payload,status,created_at,updated_at) VALUES(?,?, 'queued',?,?)`, kind, payload, now(), now())
	return err
}
func (r *Repos) CompleteJob(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE jobs SET status='done',lease_until=NULL,updated_at=? WHERE id=?`, now(), id)
	return err
}

func (r *Repos) UpdateBatchStatus(ctx context.Context, actor domain.User, id, version int64, next domain.BatchStatus) error {
	return storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var current string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM batches WHERE id=?`, id).Scan(&current); err != nil {
			return err
		}
		if !domain.BatchStatus(current).CanTransition(next) {
			return apperrors.ErrInvalidState
		}
		res, err := tx.ExecContext(ctx, `UPDATE batches SET status=?, version=version+1 WHERE id=? AND version=?`, next, id, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		return r.auditTx(ctx, tx, actor.ID, "batch", id, "status", string(next), "")
	})
}

func (r *Repos) EnsureInspection(ctx context.Context, batchID, reviewerID int64) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `INSERT INTO inspections(batch_id,reviewer_id,result,status,notes,version,created_at) VALUES(?,?, '', 'queued','',1,?)`, batchID, reviewerID, now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repos) ClaimInspection(ctx context.Context, reviewerID int64, lease time.Duration) (domain.Inspection, error) {
	var out domain.Inspection
	err := storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var id, batchID, version int64
		if err := tx.QueryRowContext(ctx, `SELECT id,batch_id,version FROM inspections WHERE status='queued' OR (status='leased' AND lease_until<?) ORDER BY id LIMIT 1`, now()).Scan(&id, &batchID, &version); err != nil {
			return apperrors.ErrNotFound
		}
		res, err := tx.ExecContext(ctx, `UPDATE inspections SET reviewer_id=?,status='leased',lease_until=?,version=version+1 WHERE id=? AND version=?`, reviewerID, time.Now().Add(lease).UTC().Format(time.RFC3339Nano), id, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		out = domain.Inspection{ID: id, BatchID: batchID, ReviewerID: reviewerID, Status: string(domain.InspectionLeased), Version: version + 1}
		return nil
	})
	return out, err
}

func (r *Repos) SubmitInspection(ctx context.Context, reviewerID, id, version int64, result, notes string) error {
	return storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE inspections SET result=?,notes=?,status=?,version=version+1,lease_until=NULL WHERE id=? AND reviewer_id=? AND status='leased' AND version=?`, result, notes, map[bool]string{true: "submitted", false: "rejected"}[result == "pass"], id, reviewerID, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		return nil
	})
}

func (r *Repos) RequeueExpiredInspections(ctx context.Context, at time.Time) (int, error) {
	res, err := r.DB.ExecContext(ctx, `UPDATE inspections SET status='queued', reviewer_id=NULL, lease_until=NULL, version=version+1 WHERE status='leased' AND lease_until<=?`, at.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *Repos) OpenComplaint(ctx context.Context, reporterID, versionID, batchID int64, region, description, key string) (domain.Complaint, error) {
	var out domain.Complaint
	err := storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		if key != "" {
			var existing string
			if err := tx.QueryRowContext(ctx, `SELECT response_json FROM idempotency_keys WHERE key=?`, key).Scan(&existing); err == nil {
				_, _ = fmt.Sscan(existing, &out.ID)
				return nil
			}
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO complaints(version_id,batch_id,reporter_id,region,description,status,version,created_at) VALUES(?,?,?,?,?,'open',1,?)`, versionID, batchID, reporterID, region, description, now())
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		out = domain.Complaint{ID: id, VersionID: versionID, BatchID: batchID, ReporterID: reporterID, Region: region, Description: description, Status: domain.CaseOpen, Version: 1}
		if key != "" {
			_, err = tx.ExecContext(ctx, `INSERT INTO idempotency_keys(key,request_hash,response_json,created_at,expires_at) VALUES(?,?,?,?,?)`, key, key, fmt.Sprint(id), now(), time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339Nano))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return out, err
}

// RegionalQuality returns a consistent, time-bounded view used by the quality
// review workflow. The same cutoff is applied to complaints and inspections so
// an audit cannot combine observations from different reporting windows.
func (r *Repos) RegionalQuality(ctx context.Context, versionID int64, since time.Time) ([]domain.RegionalQuality, error) {
	cutoff := since.UTC().Format(time.RFC3339Nano)
	rows, err := r.DB.QueryContext(ctx, `
		WITH regions AS (
			SELECT DISTINCT region FROM batches WHERE version_id=?
		), complaints_by_region AS (
			SELECT b.region, COUNT(*) AS complaints,
			       SUM(CASE WHEN c.status IN ('open','investigating') THEN 1 ELSE 0 END) AS open_cases,
			       MAX(c.created_at) AS last_complaint
			FROM batches b JOIN complaints c ON c.batch_id=b.id
			WHERE b.version_id=? AND c.created_at>=? GROUP BY b.region
		), inspections_by_region AS (
			SELECT b.region,
			       SUM(CASE WHEN i.result='fail' THEN 1 ELSE 0 END) AS failed,
			       MAX(i.created_at) AS last_inspection
			FROM batches b JOIN inspections i ON i.batch_id=b.id
			WHERE b.version_id=? AND i.created_at>=? GROUP BY b.region
		)
		SELECT r.region,
		       COALESCE(c.complaints,0), COALESCE(i.failed,0),
		       COALESCE(c.open_cases,0),
		       MAX(COALESCE(c.last_complaint,''), COALESCE(i.last_inspection,''))
		FROM regions r
		LEFT JOIN complaints_by_region c ON c.region=r.region
		LEFT JOIN inspections_by_region i ON i.region=r.region
		ORDER BY r.region`, versionID, versionID, cutoff, versionID, cutoff)
	if err != nil {
		return nil, fmt.Errorf("regional quality query: %w", err)
	}
	defer rows.Close()
	var out []domain.RegionalQuality
	for rows.Next() {
		var q domain.RegionalQuality
		var last string
		if err := rows.Scan(&q.Region, &q.ComplaintCount, &q.FailedInspections, &q.OpenCases, &last); err != nil {
			return nil, fmt.Errorf("scan regional quality: %w", err)
		}
		q.VersionID = versionID
		if last != "" {
			q.LastObservedAt, err = time.Parse(time.RFC3339Nano, last)
			if err != nil {
				return nil, fmt.Errorf("parse regional observation: %w", err)
			}
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate regional quality: %w", err)
	}
	return out, nil
}

func (r *Repos) TransitionComplaint(ctx context.Context, actorID, id, version int64, next domain.CaseStatus) error {
	return storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var current string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM complaints WHERE id=?`, id).Scan(&current); err != nil {
			return err
		}
		if !domain.CaseStatus(current).CanTransition(next) {
			return apperrors.ErrInvalidState
		}
		res, err := tx.ExecContext(ctx, `UPDATE complaints SET status=?,version=version+1 WHERE id=? AND version=?`, next, id, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		return r.auditTx(ctx, tx, actorID, "complaint", id, "status", string(next), "")
	})
}

func (r *Repos) CloseComplaintIfEvidenceArchived(ctx context.Context, actorID, id, version int64) error {
	return storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var pending int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM evidence WHERE complaint_id=? AND archived=0`, id).Scan(&pending); err != nil {
			return err
		}
		if pending > 0 {
			return apperrors.ErrInvalidState
		}
		res, err := tx.ExecContext(ctx, `UPDATE complaints SET status='closed',version=version+1 WHERE id=? AND version=? AND status='resolved'`, id, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		return r.auditTx(ctx, tx, actorID, "complaint", id, "close", "ok", "")
	})
}

func (r *Repos) AttachEvidence(ctx context.Context, actorID, complaintID int64, objectKey, hash string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO evidence(complaint_id,object_key,sha256,archived,created_at) VALUES(?,?,?,0,?)`, complaintID, objectKey, hash, now())
	return err
}
func (r *Repos) ArchiveEvidence(ctx context.Context, actorID, evidenceID int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE evidence SET archived=1 WHERE id=?`, evidenceID)
	return err
}
func (r *Repos) CreateRemediation(ctx context.Context, actorID, complaintID int64, action string, due time.Time) (domain.Remediation, error) {
	res, err := r.DB.ExecContext(ctx, `INSERT INTO remediation(complaint_id,action,status,due_at,version,created_at) VALUES(?,?, 'planned',?,1,?)`, complaintID, action, due.UTC().Format(time.RFC3339Nano), now())
	if err != nil {
		return domain.Remediation{}, err
	}
	id, _ := res.LastInsertId()
	return domain.Remediation{ID: id, ComplaintID: complaintID, Action: action, Status: string(domain.RemediationPlanned), DueAt: due, Version: 1}, nil
}
func (r *Repos) TransitionRemediation(ctx context.Context, actorID, id, version int64, next domain.RemediationStatus) error {
	return storage.WithTx(ctx, r.DB.DB, func(tx *sql.Tx) error {
		var cur string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM remediation WHERE id=?`, id).Scan(&cur); err != nil {
			return err
		}
		if !domain.RemediationStatus(cur).CanTransition(next) {
			return apperrors.ErrInvalidState
		}
		res, err := tx.ExecContext(ctx, `UPDATE remediation SET status=?,version=version+1 WHERE id=? AND version=?`, next, id, version)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return apperrors.ErrConflict
		}
		return r.auditTx(ctx, tx, actorID, "remediation", id, "status", string(next), "")
	})
}
func (r *Repos) EscalateRemediation(ctx context.Context, at time.Time) (int, error) {
	res, err := r.DB.ExecContext(ctx, `UPDATE remediation SET status='escalated',version=version+1 WHERE status='active' AND due_at<?`, at.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
func (r *Repos) CreateFulfillmentRule(ctx context.Context, actorID, versionID int64, region string, from, to time.Time) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `INSERT INTO fulfillment_rules(version_id,region,effective_from,effective_to,status,version,created_at) VALUES(?,?,?,?,'draft',1,?)`, versionID, region, from.UTC().Format(time.RFC3339Nano), to.UTC().Format(time.RFC3339Nano), now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
func (r *Repos) PublishFulfillmentRule(ctx context.Context, actorID, id, version int64) error {
	res, err := r.DB.ExecContext(ctx, `UPDATE fulfillment_rules SET status='published',version=version+1 WHERE id=? AND version=? AND status='draft'`, id, version)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return apperrors.ErrConflict
	}
	return nil
}
func (r *Repos) EnqueueNotification(ctx context.Context, recipientID int64, kind, payload string, next time.Time) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO notifications(recipient_id,kind,payload,status,attempts,next_attempt_at,created_at) VALUES(?,?,?,'queued',0,?,?)`, recipientID, kind, payload, next.UTC().Format(time.RFC3339Nano), now())
	return err
}
func (r *Repos) DeliverNotifications(ctx context.Context, limit int, at time.Time) (int, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id FROM notifications WHERE status='queued' AND next_attempt_at<=? ORDER BY id LIMIT ?`, at.UTC().Format(time.RFC3339Nano), limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		if _, err := r.DB.ExecContext(ctx, `UPDATE notifications SET status='sent',attempts=attempts+1 WHERE id=?`, id); err != nil {
			return 0, err
		}
	}
	return len(ids), rows.Err()
}
func (r *Repos) FailJob(ctx context.Context, id int64, msg string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE jobs SET status=CASE WHEN attempts+1>=5 THEN 'dead' ELSE 'queued' END, attempts=attempts+1,last_error=?,lease_until=NULL,updated_at=? WHERE id=?`, msg, now(), id)
	return err
}
func (r *Repos) auditTx(ctx context.Context, tx *sql.Tx, actorID int64, typ string, objID int64, action, result, requestID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,object_type,object_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)`, actorID, typ, objID, action, result, requestID, now())
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	return nil
}
