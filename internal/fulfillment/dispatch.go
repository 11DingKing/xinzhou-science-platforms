package fulfillment

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type DispatchRequest struct {
	SKU         string
	BatchCode   string
	Region      string
	Address     string
	RequestedAt time.Time
}
type DispatchDecision struct {
	Allowed     bool
	RuleID      int64
	Reason      string
	QualityTier string
	EvaluatedAt time.Time
}
type RuleSet struct {
	Rules       []Rule
	DefaultTier string
}

func (s RuleSet) Resolve(req DispatchRequest, now time.Time) DispatchDecision {
	region := strings.ToLower(strings.TrimSpace(req.Region))
	rules := append([]Rule(nil), s.Rules...)
	sort.SliceStable(rules, func(i, j int) bool { return rules[i].EffectiveFrom.After(rules[j].EffectiveFrom) })
	for _, rule := range rules {
		if rule.Region == region && rule.Status == "published" && !now.Before(rule.EffectiveFrom) && now.Before(rule.EffectiveTo) {
			return DispatchDecision{Allowed: true, RuleID: rule.ID, QualityTier: "declared", EvaluatedAt: now.UTC()}
		}
	}
	if s.DefaultTier == "" {
		return DispatchDecision{Allowed: false, Reason: "no active rule", EvaluatedAt: now.UTC()}
	}
	return DispatchDecision{Allowed: true, QualityTier: s.DefaultTier, EvaluatedAt: now.UTC()}
}
func (s RuleSet) Validate() error {
	if s.DefaultTier == "" && len(s.Rules) == 0 {
		return errors.New("empty rule set")
	}
	for _, r := range s.Rules {
		if r.ID < 1 || r.Region == "" || !r.EffectiveTo.After(r.EffectiveFrom) {
			return fmt.Errorf("invalid rule %d", r.ID)
		}
	}
	return nil
}
func CompareRequests(a, b DispatchRequest) bool {
	return strings.EqualFold(a.SKU, b.SKU) && strings.EqualFold(a.BatchCode, b.BatchCode) && strings.EqualFold(a.Region, b.Region)
}
func RequireConsistentTier(decisions []DispatchDecision) error {
	if len(decisions) < 2 {
		return nil
	}
	tier := decisions[0].QualityTier
	for _, d := range decisions[1:] {
		if d.QualityTier != tier {
			return errors.New("quality tier differs")
		}
	}
	return nil
}
