package diagnostic

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDiagnosticReport(t *testing.T) {
	r := Run(context.Background(), []Check{{Name: "db", Critical: true, Run: func(context.Context) error { return nil }}, {Name: "queue", Critical: false, Run: func(context.Context) error { return errors.New("down") }}}, time.Now())
	if r.Ready || Healthy(r) || len(Failed(r)) != 1 || !strings.Contains(Summary(r), "1/2") {
		t.Fatalf("report=%+v", r)
	}
	empty := Run(context.Background(), nil, time.Now())
	if empty.Ready {
		t.Fatal("empty ready")
	}
}
