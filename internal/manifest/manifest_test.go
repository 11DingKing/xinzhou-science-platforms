package manifest

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseManifest(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	input := strings.NewReader(`{"sku":"s1","batch_code":"b1","region":"City","declared_at":"` + now + `"}` + "\n" + `{"sku":"","batch_code":"b2","region":"x","declared_at":"` + now + `"}` + "\n" + `not-json` + "\n")
	rows, report, err := ParseJSONLines(context.Background(), input, 10)
	if err != nil || len(rows) != 1 || report.Accepted != 1 || report.Rejected != 2 {
		t.Fatalf("rows=%+v report=%+v err=%v", rows, report, err)
	}
	if len(GroupByRegion(rows)["city"]) != 1 {
		t.Fatal("group")
	}
}
func TestManifestHelpers(t *testing.T) {
	now := time.Now()
	rows := []Row{{SKU: "s", BatchCode: "b2", Region: "city", DeclaredAt: now}, {SKU: "s", BatchCode: "b1", Region: "city", DeclaredAt: now}, {SKU: "s", BatchCode: "b1", Region: "city", DeclaredAt: now}}
	if len(Dedupe(rows)) != 2 {
		t.Fatal("dedupe")
	}
	if SortStable(rows)[0].BatchCode != "b1" {
		t.Fatal("sort")
	}
	if len(MissingRegions(rows, []string{"city", "county"})) != 1 {
		t.Fatal("missing")
	}
	if Validate(Row{}) == nil {
		t.Fatal("empty accepted")
	}
}
func TestParseHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := ParseJSONLines(ctx, strings.NewReader(`{"sku":"s","batch_code":"b","region":"r","declared_at":"2026-01-01T00:00:00Z"}`), 10)
	if err == nil {
		t.Fatal("cancel ignored")
	}
}
