package risk

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"sort"
	"time"
)

type Signal struct {
	Kind       string
	Weight     int
	ObservedAt time.Time
	Region     string
	BatchID    int64
}
type Profile struct {
	BatchID   int64
	Score     int
	Level     string
	Signals   []Signal
	UpdatedAt time.Time
}
type Calculator struct {
	Weights map[string]int
	Now     func() time.Time
}

func NewCalculator() Calculator {
	return Calculator{Weights: map[string]int{"material_mismatch": 40, "configuration_mismatch": 30, "region_mismatch": 20, "late_evidence": 15, "repeat_complaint": 25}, Now: time.Now}
}
func (c Calculator) Score(signals []Signal) Profile {
	score := 0
	copySignals := append([]Signal(nil), signals...)
	for _, s := range copySignals {
		if w, ok := c.Weights[s.Kind]; ok {
			score += w
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
	var id int64
	if len(copySignals) > 0 {
		id = copySignals[0].BatchID
	}
	return Profile{BatchID: id, Score: score, Level: level, Signals: copySignals, UpdatedAt: c.Now().UTC()}
}
func Rank(profiles []Profile) []Profile {
	out := append([]Profile(nil), profiles...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].BatchID < out[j].BatchID
		}
		return out[i].Score > out[j].Score
	})
	return out
}
func IsActionable(p Profile) bool { return p.Score >= 50 }
func Explain(p Profile) []string {
	out := []string{}
	for _, s := range p.Signals {
		out = append(out, s.Kind)
	}
	return out
}
func SupportsBatch(p Profile, b domain.Batch) bool {
	return p.BatchID == b.ID && b.Status != domain.BatchArchived
}
