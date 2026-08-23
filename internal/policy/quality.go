package policy

import (
	"errors"
	"fmt"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"strings"
	"time"
)

var (
	ErrMissingDisclosure = errors.New("required disclosure is missing")
	ErrQualityMismatch   = errors.New("quality evidence does not match declaration")
	ErrWindowClosed      = errors.New("quality review window is closed")
)

type Declaration struct {
	SKU, Channel, Region, Material, Configuration string
	DeclaredAt                                    time.Time
	ExpiresAt                                     time.Time
}
type Sample struct {
	SKU, Channel, Region, Material, Configuration string
	ObservedAt                                    time.Time
	Score                                         int
	Source                                        string
}
type Decision struct {
	Allowed  bool
	Reasons  []string
	Severity string
}

func (p Declaration) Validate(now time.Time) error {
	if strings.TrimSpace(p.SKU) == "" || strings.TrimSpace(p.Channel) == "" || strings.TrimSpace(p.Region) == "" {
		return ErrMissingDisclosure
	}
	if p.ExpiresAt.Before(now) {
		return ErrWindowClosed
	}
	if p.ExpiresAt.Before(p.DeclaredAt) {
		return errors.New("declaration expiry precedes creation")
	}
	return nil
}
func Compare(d Declaration, s Sample) Decision {
	reasons := []string{}
	if d.SKU != s.SKU {
		reasons = append(reasons, "sku differs")
	}
	if d.Channel != s.Channel {
		reasons = append(reasons, "channel differs")
	}
	if d.Region != s.Region {
		reasons = append(reasons, "region differs")
	}
	if d.Material != s.Material {
		reasons = append(reasons, "material differs")
	}
	if d.Configuration != s.Configuration {
		reasons = append(reasons, "configuration differs")
	}
	severity := "none"
	if len(reasons) > 0 {
		severity = "review"
	}
	if s.Score < 60 {
		reasons = append(reasons, "score below release threshold")
		severity = "critical"
	}
	return Decision{Allowed: len(reasons) == 0, Reasons: reasons, Severity: severity}
}
func RequirePublished(v domain.ProductVersion) error {
	if v.Status != domain.VersionPublished {
		return fmt.Errorf("version %d is not published", v.ID)
	}
	return nil
}
func RequireBatchReviewable(b domain.Batch, now time.Time) error {
	if b.Status != domain.BatchSampling {
		return fmt.Errorf("batch %d is not in sampling", b.ID)
	}
	if b.ExpiresAt.Before(now) {
		return ErrWindowClosed
	}
	return nil
}
func EscalationLevel(score int) string {
	switch {
	case score < 40:
		return "critical"
	case score < 60:
		return "high"
	case score < 80:
		return "medium"
	default:
		return "low"
	}
}
func NormalizeRegion(v string) string { return strings.ToLower(strings.Join(strings.Fields(v), " ")) }
func MergeReasons(parts ...[]string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, part := range parts {
		for _, value := range part {
			if value != "" && !seen[value] {
				seen[value] = true
				out = append(out, value)
			}
		}
	}
	return out
}
