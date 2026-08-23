package export

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestExportFormats(t *testing.T) {
	records := []Record{{ID: 1, SKU: "s", BatchCode: "b", Region: "city", Status: "flagged", Score: 80, At: time.Now()}}
	var b bytes.Buffer
	if err := JSON(&b, records); err != nil || !strings.Contains(b.String(), `"sku":"s"`) {
		t.Fatal("json")
	}
	b.Reset()
	if err := CSV(&b, records); err != nil || !strings.Contains(b.String(), "batch_code") {
		t.Fatal("csv")
	}
	b.Reset()
	if err := Lines(&b, records); err != nil || !strings.Contains(b.String(), "flagged") {
		t.Fatal("lines")
	}
	if len(Filter(records, "flagged", 70)) != 1 || len(Filter(records, "open", 0)) != 0 {
		t.Fatal("filter")
	}
}
