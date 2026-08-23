package dispute

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusOpen      Status = "open"
	StatusMediating Status = "mediating"
	StatusResolved  Status = "resolved"
	StatusEscalated Status = "escalated"
	StatusClosed    Status = "closed"
)

type Dispute struct {
	ID          int64
	ComplaintID int64
	MerchantID  int64
	ReviewerID  int64
	Status      Status
	Reason      string
	Messages    []Message
	OpenedAt    time.Time
	UpdatedAt   time.Time
}
type Message struct {
	ID       int64
	AuthorID int64
	Body     string
	At       time.Time
	Private  bool
}

func Open(complaintID, merchantID int64, reason string, now time.Time) (Dispute, error) {
	if complaintID < 1 || merchantID < 1 || strings.TrimSpace(reason) == "" {
		return Dispute{}, errors.New("invalid dispute")
	}
	return Dispute{ComplaintID: complaintID, MerchantID: merchantID, Status: StatusOpen, Reason: reason, OpenedAt: now.UTC(), UpdatedAt: now.UTC()}, nil
}
func Assign(d *Dispute, reviewerID int64, now time.Time) error {
	if d.Status != StatusOpen || reviewerID < 1 {
		return errors.New("dispute not assignable")
	}
	d.ReviewerID = reviewerID
	d.Status = StatusMediating
	d.UpdatedAt = now.UTC()
	return nil
}
func AddMessage(d *Dispute, m Message, now time.Time) error {
	if d.Status != StatusMediating {
		return errors.New("dispute not mediating")
	}
	if strings.TrimSpace(m.Body) == "" {
		return errors.New("empty message")
	}
	m.At = now.UTC()
	d.Messages = append(d.Messages, m)
	d.UpdatedAt = m.At
	return nil
}
func Resolve(d *Dispute, now time.Time) error {
	if d.Status != StatusMediating || len(d.Messages) == 0 {
		return errors.New("mediation incomplete")
	}
	d.Status = StatusResolved
	d.UpdatedAt = now.UTC()
	return nil
}
func Escalate(d *Dispute, now time.Time) error {
	if d.Status != StatusOpen && d.Status != StatusMediating {
		return errors.New("dispute cannot escalate")
	}
	d.Status = StatusEscalated
	d.UpdatedAt = now.UTC()
	return nil
}
func Close(d *Dispute, now time.Time) error {
	if d.Status != StatusResolved && d.Status != StatusEscalated {
		return errors.New("dispute not final")
	}
	d.Status = StatusClosed
	d.UpdatedAt = now.UTC()
	return nil
}
func VisibleMessages(d Dispute, viewerID int64, reviewer bool) []Message {
	out := []Message{}
	for _, m := range d.Messages {
		if !m.Private || reviewer || m.AuthorID == viewerID {
			out = append(out, m)
		}
	}
	return out
}
