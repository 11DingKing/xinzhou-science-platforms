package analytics

import (
	"sort"
	"time"
)

type Observation struct {
	Region  string
	SKU     string
	BatchID int64
	Score   int
	At      time.Time
}
type RegionSummary struct {
	Region       string
	Samples      int
	AverageScore float64
	LowestScore  int
	HighestScore int
	Flagged      bool
}
type Summary struct {
	ByRegion       []RegionSummary
	TotalSamples   int
	FlaggedRegions int
	GeneratedAt    time.Time
}

func Build(observations []Observation, now time.Time) Summary {
	groups := map[string][]Observation{}
	for _, o := range observations {
		groups[o.Region] = append(groups[o.Region], o)
	}
	out := Summary{GeneratedAt: now.UTC(), TotalSamples: len(observations)}
	for region, items := range groups {
		sum := 0
		low := items[0].Score
		high := items[0].Score
		for _, item := range items {
			sum += item.Score
			if item.Score < low {
				low = item.Score
			}
			if item.Score > high {
				high = item.Score
			}
		}
		avg := float64(sum) / float64(len(items))
		flagged := avg < 70 || low < 50
		if flagged {
			out.FlaggedRegions++
		}
		out.ByRegion = append(out.ByRegion, RegionSummary{Region: region, Samples: len(items), AverageScore: avg, LowestScore: low, HighestScore: high, Flagged: flagged})
	}
	sort.Slice(out.ByRegion, func(i, j int) bool { return out.ByRegion[i].Region < out.ByRegion[j].Region })
	return out
}
func Compare(a, b Summary) map[string]float64 {
	out := map[string]float64{}
	for _, left := range a.ByRegion {
		for _, right := range b.ByRegion {
			if left.Region == right.Region {
				out[left.Region] = left.AverageScore - right.AverageScore
			}
		}
	}
	return out
}
func Worst(summary Summary, n int) []RegionSummary {
	out := append([]RegionSummary(nil), summary.ByRegion...)
	sort.Slice(out, func(i, j int) bool { return out[i].AverageScore < out[j].AverageScore })
	if n < 0 {
		n = 0
	}
	if n > len(out) {
		n = len(out)
	}
	return out[:n]
}
