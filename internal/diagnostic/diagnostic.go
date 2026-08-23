package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Check struct {
	Name     string
	Run      func(context.Context) error
	Critical bool
}
type Result struct {
	Name     string
	OK       bool
	Error    string
	Duration time.Duration
}
type Report struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Results    []Result
	Ready      bool
}

func Run(ctx context.Context, checks []Check, now time.Time) Report {
	report := Report{StartedAt: now.UTC()}
	for _, check := range checks {
		start := time.Now()
		err := error(nil)
		if check.Run == nil {
			err = errors.New("check has no runner")
		} else {
			err = check.Run(ctx)
		}
		result := Result{Name: check.Name, OK: err == nil, Duration: time.Since(start)}
		if err != nil {
			result.Error = err.Error()
		}
		report.Results = append(report.Results, result)
		if err != nil && check.Critical {
			report.Ready = false
		} else if report.Ready == false && len(report.Results) == 1 {
			report.Ready = true
		}
	}
	if len(report.Results) == 0 {
		report.Ready = false
	}
	for _, result := range report.Results {
		if !result.OK {
			report.Ready = false
		}
	}
	report.FinishedAt = time.Now().UTC()
	return report
}
func Summary(r Report) string {
	ok := 0
	for _, result := range r.Results {
		if result.OK {
			ok++
		}
	}
	return fmt.Sprintf("%d/%d checks passed", ok, len(r.Results))
}
func Failed(r Report) []Result {
	out := []Result{}
	for _, result := range r.Results {
		if !result.OK {
			out = append(out, result)
		}
	}
	return out
}
func Healthy(r Report) bool { return r.Ready && len(Failed(r)) == 0 }
