package dispute

import (
	"testing"
	"time"
)

func TestDisputeLifecycle(t *testing.T) {
	now := time.Now()
	d, err := Open(1, 2, "quality differs", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := Assign(&d, 3, now); err != nil {
		t.Fatal(err)
	}
	if err := AddMessage(&d, Message{AuthorID: 3, Body: "review started"}, now); err != nil {
		t.Fatal(err)
	}
	if err := Resolve(&d, now); err != nil {
		t.Fatal(err)
	}
	if err := Close(&d, now); err != nil {
		t.Fatal(err)
	}
	if len(VisibleMessages(d, 2, false)) != 1 {
		t.Fatal("visibility")
	}
}
func TestDisputeEscalation(t *testing.T) {
	d, err := Open(1, 2, "late response", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := Escalate(&d, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := Close(&d, time.Now()); err != nil {
		t.Fatal(err)
	}
}
