package lifecycle

import (
	"testing"
	"time"
)

func TestLifecycleManager(t *testing.T) {
	now := time.Now()
	m := New(now)
	if err := m.Transition(StateRunning, now.Add(time.Second), "start"); err != nil {
		t.Fatal(err)
	}
	if err := m.RequireRunning(); err != nil {
		t.Fatal(err)
	}
	if m.Uptime(now.Add(2*time.Second)) != time.Second {
		t.Fatal("uptime")
	}
	if err := m.Transition(StateStopped, now, "bad skip"); err == nil {
		t.Fatal("skip accepted")
	}
	if err := m.Transition(StateStopping, now.Add(2*time.Second), "shutdown"); err != nil {
		t.Fatal(err)
	}
	if err := m.Transition(StateStopped, now.Add(3*time.Second), "done"); err != nil {
		t.Fatal(err)
	}
}
