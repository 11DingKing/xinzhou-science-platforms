package investigation

import (
	"errors"
	"fmt"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"sort"
	"strings"
	"time"
)

type Finding struct {
	ID          int64
	CaseID      int64
	Kind        string
	Severity    string
	Statement   string
	EvidenceIDs []int64
	CreatedAt   time.Time
}
type Case struct {
	ID          int64
	ComplaintID int64
	Status      domain.CaseStatus
	OwnerID     int64
	Findings    []Finding
	OpenedAt    time.Time
	ClosedAt    *time.Time
}
type Assignment struct {
	ReviewerID int64
	AssignedAt time.Time
	DueAt      time.Time
	Status     string
}

func Start(c Case, actorID int64, now time.Time) (Case, error) {
	if c.Status != domain.CaseOpen {
		return c, errors.New("case is not open")
	}
	c.Status = domain.CaseInvestigating
	c.OwnerID = actorID
	if c.OpenedAt.IsZero() {
		c.OpenedAt = now.UTC()
	}
	return c, nil
}
func AddFinding(c *Case, f Finding) error {
	if c.Status != domain.CaseInvestigating {
		return errors.New("case not investigating")
	}
	if strings.TrimSpace(f.Statement) == "" || len(f.EvidenceIDs) == 0 {
		return errors.New("finding requires statement and evidence")
	}
	f.CaseID = c.ID
	c.Findings = append(c.Findings, f)
	return nil
}
func Resolve(c Case, now time.Time) (Case, error) {
	if c.Status != domain.CaseInvestigating || len(c.Findings) == 0 {
		return c, errors.New("case lacks findings")
	}
	c.Status = domain.CaseResolved
	return c, nil
}
func Close(c Case, now time.Time) (Case, error) {
	if c.Status != domain.CaseResolved {
		return c, errors.New("case not resolved")
	}
	c.Status = domain.CaseClosed
	t := now.UTC()
	c.ClosedAt = &t
	return c, nil
}
func Assign(c Case, reviewerID int64, now time.Time, ttl time.Duration) (Assignment, error) {
	if c.Status != domain.CaseInvestigating {
		return Assignment{}, errors.New("case not investigating")
	}
	if reviewerID < 1 {
		return Assignment{}, errors.New("reviewer required")
	}
	return Assignment{ReviewerID: reviewerID, AssignedAt: now.UTC(), DueAt: now.UTC().Add(ttl), Status: "active"}, nil
}
func Severity(f []Finding) string {
	best := 0
	for _, item := range f {
		value := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}[strings.ToLower(item.Severity)]
		if value > best {
			best = value
		}
	}
	for key, value := range map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1} {
		if value == best {
			return key
		}
	}
	return "none"
}
func SortFindings(f []Finding) []Finding {
	out := append([]Finding(nil), f...)
	level := func(value string) int {
		return map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}[strings.ToLower(value)]
	}
	sort.SliceStable(out, func(i, j int) bool {
		if level(out[i].Severity) == level(out[j].Severity) {
			return out[i].ID < out[j].ID
		}
		return level(out[i].Severity) > level(out[j].Severity)
	})
	return out
}
func Explain(c Case) string {
	return fmt.Sprintf("case %d status=%s findings=%d", c.ID, c.Status, len(c.Findings))
}
