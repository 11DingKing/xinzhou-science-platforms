package validator

import (
	"testing"
	"time"
)

func TestValidators(t *testing.T) {
	if !Email("quality@example.test") || Email("not email") {
		t.Fatal("email")
	}
	if !BatchCode("BATCH_2026") || BatchCode("bad code") {
		t.Fatal("batch code")
	}
	if !Region("county") || Region("") {
		t.Fatal("region")
	}
	base := time.Now()
	if !Window(base, base.Add(time.Hour)) || Window(base.Add(time.Hour), base) {
		t.Fatal("window")
	}
	if Required("ok", "") == nil || Required("a", "b") != nil {
		t.Fatal("required")
	}
	if !Unique([]string{"a", "b"}) || Unique([]string{"a", "a"}) {
		t.Fatal("unique")
	}
}
