package redaction

import (
	"strings"
	"testing"
)

func TestRedactsSensitiveValues(t *testing.T) {
	got := Text("contact user@example.com with Bearer abc123")
	if got == "contact user@example.com with Bearer abc123" || !strings.Contains(got, "***") {
		t.Fatalf("redacted=%s", got)
	}
	fields := Fields(map[string]string{"password": "secret", "note": "user@example.com"})
	if fields["password"] != "***" || fields["note"] == "user@example.com" {
		t.Fatal("fields")
	}
}
