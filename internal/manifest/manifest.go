package manifest

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var ErrInvalidRow = errors.New("invalid manifest row")

type Row struct {
	SKU           string    `json:"sku"`
	BatchCode     string    `json:"batch_code"`
	Region        string    `json:"region"`
	Material      string    `json:"material"`
	Configuration string    `json:"configuration"`
	DeclaredAt    time.Time `json:"declared_at"`
}
type Report struct {
	Accepted, Rejected int
	Errors             []string
}

func ParseJSONLines(ctx context.Context, r io.Reader, limit int) ([]Row, Report, error) {
	if limit < 1 {
		limit = 1000
	}
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 4096), 1024*1024)
	rows := []Row{}
	report := Report{}
	line := 0
	for scan.Scan() {
		line++
		select {
		case <-ctx.Done():
			return rows, report, ctx.Err()
		default:
		}
		if len(rows) >= limit {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d exceeds limit", line))
			continue
		}
		var row Row
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d: %v", line, err))
			continue
		}
		if err := Validate(row); err != nil {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d: %v", line, err))
			continue
		}
		rows = append(rows, row)
		report.Accepted++
	}
	if err := scan.Err(); err != nil {
		return rows, report, err
	}
	return rows, report, nil
}
func Validate(r Row) error {
	if strings.TrimSpace(r.SKU) == "" || strings.TrimSpace(r.BatchCode) == "" || strings.TrimSpace(r.Region) == "" {
		return ErrInvalidRow
	}
	if r.DeclaredAt.IsZero() {
		return fmt.Errorf("declared_at is required")
	}
	return nil
}
func GroupByRegion(rows []Row) map[string][]Row {
	out := map[string][]Row{}
	for _, row := range rows {
		key := strings.ToLower(strings.TrimSpace(row.Region))
		out[key] = append(out[key], row)
	}
	return out
}
func Dedupe(rows []Row) []Row {
	seen := map[string]bool{}
	out := []Row{}
	for _, row := range rows {
		key := row.SKU + "|" + row.BatchCode + "|" + row.Region
		if !seen[key] {
			seen[key] = true
			out = append(out, row)
		}
	}
	return out
}
func MissingRegions(rows []Row, required []string) []string {
	seen := map[string]bool{}
	for _, r := range rows {
		seen[strings.ToLower(r.Region)] = true
	}
	out := []string{}
	for _, region := range required {
		if !seen[strings.ToLower(region)] {
			out = append(out, region)
		}
	}
	return out
}
func SortStable(rows []Row) []Row {
	out := append([]Row(nil), rows...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].BatchCode < out[j-1].BatchCode; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
