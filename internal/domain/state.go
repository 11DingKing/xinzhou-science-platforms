package domain

import (
	"fmt"
	"time"
)

type InspectionStatus string

const (
	InspectionQueued    InspectionStatus = "queued"
	InspectionLeased    InspectionStatus = "leased"
	InspectionSubmitted InspectionStatus = "submitted"
	InspectionRejected  InspectionStatus = "rejected"
)

type EvidenceStatus string

const (
	EvidencePending  EvidenceStatus = "pending"
	EvidenceArchived EvidenceStatus = "archived"
)

type RemediationStatus string

const (
	RemediationPlanned   RemediationStatus = "planned"
	RemediationActive    RemediationStatus = "active"
	RemediationDone      RemediationStatus = "done"
	RemediationEscalated RemediationStatus = "escalated"
)

func (s VersionStatus) CanTransition(to VersionStatus) bool {
	return (s == VersionDraft && to == VersionPublished) || (s == VersionPublished && to == VersionWithdrawn)
}
func (s BatchStatus) CanTransition(to BatchStatus) bool {
	return (s == BatchPending && to == BatchSampling) || (s == BatchSampling && (to == BatchCleared || to == BatchFlagged)) || (s == BatchFlagged && to == BatchArchived) || (s == BatchCleared && to == BatchArchived)
}
func (s CaseStatus) CanTransition(to CaseStatus) bool {
	return (s == CaseOpen && to == CaseInvestigating) || (s == CaseInvestigating && to == CaseResolved) || (s == CaseResolved && to == CaseClosed)
}
func (s RemediationStatus) CanTransition(to RemediationStatus) bool {
	return (s == RemediationPlanned && to == RemediationActive) || (s == RemediationActive && (to == RemediationDone || to == RemediationEscalated))
}

func RequireTransition(from, to string, allowed bool) error {
	if !allowed {
		return fmt.Errorf("invalid transition %s -> %s", from, to)
	}
	return nil
}
func WithinWindow(now, deadline time.Time) bool { return !now.After(deadline) }
func LeaseExpired(now, leaseUntil time.Time) bool {
	return leaseUntil.IsZero() || !now.Before(leaseUntil)
}
