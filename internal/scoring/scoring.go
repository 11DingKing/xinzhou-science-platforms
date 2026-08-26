package scoring

import (
	"errors"
	"sort"
	"strings"
)

type Rule struct {
	Name      string
	Weight    int
	Threshold int
	Enabled   bool
}
type Input struct {
	MaterialMatch      bool
	ConfigurationMatch bool
	RegionMatch        bool
	ComplaintCount     int
	InspectionScore    int
}
type Result struct {
	Score     int
	Level     string
	Reasons   []string
	RuleNames []string
}
type Engine struct{ Rules []Rule }

func DefaultEngine() Engine {
	return Engine{Rules: []Rule{{"material", 30, 1, true}, {"configuration", 25, 1, true}, {"region", 20, 1, true}, {"complaint", 30, 1, true}, {"inspection", 25, 70, true}}}
}
func (e Engine) Evaluate(in Input) Result {
	score := 0
	reasons := []string{}
	names := []string{}
	for _, r := range e.Rules {
		if !r.Enabled {
			continue
		}
		trigger := false
		reason := r.Name
		switch r.Name {
		case "material":
			trigger = !in.MaterialMatch
		case "configuration":
			trigger = !in.ConfigurationMatch
		case "region":
			trigger = !in.RegionMatch
		case "complaint":
			trigger = in.ComplaintCount >= r.Threshold
		case "inspection":
			trigger = in.InspectionScore < r.Threshold
		}
		if trigger {
			score += r.Weight
			reasons = append(reasons, reason)
			names = append(names, r.Name)
		}
	}
	if score > 100 {
		score = 100
	}
	level := "low"
	if score >= 80 {
		level = "critical"
	} else if score >= 50 {
		level = "high"
	} else if score >= 25 {
		level = "medium"
	}
	return Result{Score: score, Level: level, Reasons: reasons, RuleNames: names}
}
func (e Engine) Validate() error {
	seen := map[string]bool{}
	for _, r := range e.Rules {
		if strings.TrimSpace(r.Name) == "" || r.Weight < 1 || seen[r.Name] {
			return errors.New("invalid score rule")
		}
		seen[r.Name] = true
	}
	return nil
}
func SortRules(rules []Rule) []Rule {
	out := append([]Rule(nil), rules...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Weight > out[j].Weight })
	return out
}
func Explain(r Result) string { return strings.Join(r.Reasons, ", ") }
