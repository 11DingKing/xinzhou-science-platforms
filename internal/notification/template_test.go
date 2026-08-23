package notification

import (
	"testing"
)

func TestNotificationTemplates(t *testing.T) {
	tpl := DefaultTemplates()["complaint"]
	text, err := Render(tpl, map[string]string{"batch": "B1", "region": "county"})
	if err != nil || text == "" {
		t.Fatalf("render=%q err=%v", text, err)
	}
	if _, err := Render(tpl, map[string]string{"batch": "B1"}); err == nil {
		t.Fatal("missing value accepted")
	}
	if !IsKnown("remediation") || IsKnown("unknown") {
		t.Fatal("known")
	}
}
