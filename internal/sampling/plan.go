package sampling

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"sort"
	"strings"
	"time"
)

var ErrNoRegions = errors.New("at least one sampling region is required")
var ErrInvalidQuota = errors.New("sampling quota must be positive")

type RegionQuota struct {
	Region   string
	Quota    int
	Priority int
}
type Plan struct {
	ID        string
	VersionID int64
	BatchID   int64
	Regions   []RegionQuota
	CreatedAt time.Time
	ExpiresAt time.Time
	Status    string
}
type Candidate struct {
	Region      string
	AddressHash string
	Priority    int
	Selected    bool
}

func BuildPlan(versionID, batchID int64, regions []RegionQuota, now time.Time, ttl time.Duration) (Plan, error) {
	if len(regions) == 0 {
		return Plan{}, ErrNoRegions
	}
	copyRegions := append([]RegionQuota(nil), regions...)
	for i := range copyRegions {
		copyRegions[i].Region = strings.ToLower(strings.TrimSpace(copyRegions[i].Region))
		if copyRegions[i].Region == "" || copyRegions[i].Quota < 1 {
			return Plan{}, ErrInvalidQuota
		}
	}
	sort.Slice(copyRegions, func(i, j int) bool {
		if copyRegions[i].Priority == copyRegions[j].Priority {
			return copyRegions[i].Region < copyRegions[j].Region
		}
		return copyRegions[i].Priority > copyRegions[j].Priority
	})
	raw := fmt.Sprintf("%d:%d:%v", versionID, batchID, copyRegions)
	sum := sha256.Sum256([]byte(raw))
	return Plan{ID: hex.EncodeToString(sum[:12]), VersionID: versionID, BatchID: batchID, Regions: copyRegions, CreatedAt: now.UTC(), ExpiresAt: now.UTC().Add(ttl), Status: "draft"}, nil
}
func Publish(p Plan, now time.Time) (Plan, error) {
	if p.Status != "draft" {
		return p, errors.New("plan is not draft")
	}
	if !now.Before(p.ExpiresAt) {
		return p, errors.New("plan expired")
	}
	p.Status = "published"
	return p, nil
}
func SelectCandidates(p Plan, candidates []Candidate) []Candidate {
	out := []Candidate{}
	byRegion := map[string]int{}
	for _, c := range candidates {
		byRegion[c.Region]++
	}
	for _, quota := range p.Regions {
		need := quota.Quota
		for _, c := range candidates {
			if c.Region != quota.Region || need == 0 {
				continue
			}
			c.Selected = true
			out = append(out, c)
			need--
		}
	}
	_ = byRegion
	return out
}
func ValidatePlan(p Plan, b domain.Batch, now time.Time) error {
	if p.BatchID != b.ID || p.VersionID == 0 {
		return errors.New("plan target mismatch")
	}
	if b.Status != domain.BatchSampling {
		return errors.New("batch is not sampling")
	}
	if !now.Before(p.ExpiresAt) {
		return errors.New("plan expired")
	}
	return nil
}
func Regions(p Plan) []string {
	out := make([]string, 0, len(p.Regions))
	for _, r := range p.Regions {
		out = append(out, r.Region)
	}
	return out
}
