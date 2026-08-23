package domain

import "time"

type Role string

const (
	RoleReviewer      Role = "reviewer"
	RoleMerchant      Role = "merchant"
	RolePlatformAdmin Role = "platform_admin"
	RolePlatformLead  Role = "platform_lead"
)

type VersionStatus string

const (
	VersionDraft     VersionStatus = "draft"
	VersionPublished VersionStatus = "published"
	VersionWithdrawn VersionStatus = "withdrawn"
)

type BatchStatus string

const (
	BatchPending  BatchStatus = "pending"
	BatchSampling BatchStatus = "sampling"
	BatchCleared  BatchStatus = "cleared"
	BatchFlagged  BatchStatus = "flagged"
	BatchArchived BatchStatus = "archived"
)

type CaseStatus string

const (
	CaseOpen          CaseStatus = "open"
	CaseInvestigating CaseStatus = "investigating"
	CaseResolved      CaseStatus = "resolved"
	CaseClosed        CaseStatus = "closed"
)

type User struct {
	ID                  int64
	Email, PasswordHash string
	Role                Role
	Disabled            bool
}
type ProductVersion struct {
	ID                        int64
	MerchantID                int64
	SKU, DisplayName, Channel string
	Status                    VersionStatus
	CreatedAt                 time.Time
	Version                   int64
}
type Batch struct {
	ID, VersionID int64
	Code          string
	Region        string
	Status        BatchStatus
	ExpiresAt     time.Time
	Version       int64
}
type Inspection struct {
	ID, BatchID, ReviewerID int64
	Result                  string
	Status                  string
	Notes                   string
	Version                 int64
}
type Complaint struct {
	ID, VersionID, BatchID, ReporterID int64
	Region, Description                string
	Status                             CaseStatus
	Version                            int64
}
type Evidence struct {
	ID, ComplaintID, InspectionID int64
	ObjectKey, Sha256             string
	Archived                      bool
	CreatedAt                     time.Time
}
type Remediation struct {
	ID, ComplaintID int64
	Action, Status  string
	DueAt           time.Time
	Version         int64
}

// RegionalQuality summarizes the evidence that a product version has behaved
// differently across fulfillment regions during an audit window.
type RegionalQuality struct {
	VersionID         int64
	Region            string
	ComplaintCount    int
	FailedInspections int
	OpenCases         int
	LastObservedAt    time.Time
}

func (q RegionalQuality) RequiresInvestigation(threshold int) bool {
	if threshold < 1 {
		threshold = 1
	}
	return q.FailedInspections >= threshold || q.OpenCases >= threshold
}

type AuditEvent struct {
	ID                        int64
	ActorID                   int64
	ObjectType                string
	ObjectID                  int64
	Action, Result, RequestID string
	CreatedAt                 time.Time
}
type Notification struct {
	ID                    int64
	RecipientID           int64
	Kind, Payload, Status string
	Attempts              int
	NextAttemptAt         time.Time
}
