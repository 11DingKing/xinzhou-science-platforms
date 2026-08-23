package appeal

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusReviewing Status = "reviewing"
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
)

type Appeal struct {
	ID          int64
	CaseID      int64
	MerchantID  int64
	Reason      string
	EvidenceIDs []int64
	Status      Status
	Version     int64
	CreatedAt   time.Time
}

func Submit(a Appeal, now time.Time) (Appeal, error) {
	if a.Status != StatusDraft || strings.TrimSpace(a.Reason) == "" || len(a.EvidenceIDs) == 0 {
		return a, errors.New("appeal incomplete")
	}
	a.Status = StatusSubmitted
	a.CreatedAt = now.UTC()
	a.Version++
	return a, nil
}
func StartReview(a Appeal, reviewerID int64) (Appeal, error) {
	if a.Status != StatusSubmitted || reviewerID < 1 {
		return a, errors.New("appeal not ready")
	}
	a.Status = StatusReviewing
	return a, nil
}
func Decide(a Appeal, accepted bool) (Appeal, error) {
	if a.Status != StatusReviewing {
		return a, errors.New("appeal not reviewing")
	}
	if accepted {
		a.Status = StatusAccepted
	} else {
		a.Status = StatusRejected
	}
	a.Version++
	return a, nil
}
func IsFinal(s Status) bool    { return s == StatusAccepted || s == StatusRejected }
func CanEdit(a Appeal) bool    { return a.Status == StatusDraft }
func ReasonWords(a Appeal) int { return len(strings.Fields(a.Reason)) }
