package operation

import (
	"testing"
	"time"
)

func TestOperationLifecycle(t *testing.T) {
	s := New()
	if _, err := s.Create("x", "export", 2); err != nil {
		t.Fatal(err)
	}
	op, err := s.Start("x", time.Now())
	if err != nil || op.State != StateRunning {
		t.Fatal("start")
	}
	if err := s.Update("x", 50, "half"); err != nil {
		t.Fatal(err)
	}
	if err := s.Update("x", 40, "back"); err == nil {
		t.Fatal("backward")
	}
	if err := s.Finish("x", StateSucceeded, time.Now()); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get("x"); got.State != StateSucceeded || got.Progress != 100 {
		t.Fatal("finish")
	}
}
