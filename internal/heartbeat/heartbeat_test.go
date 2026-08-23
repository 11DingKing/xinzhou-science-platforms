package heartbeat

import (
	"testing"
	"time"
)

func TestHeartbeatExpiry(t *testing.T) {
	now := time.Now()
	m := New(time.Minute)
	m.Beat("worker", "ok", now)
	if !m.Healthy("worker", now.Add(time.Second)) || !m.AllHealthy(now) {
		t.Fatal("healthy")
	}
	if m.Healthy("worker", now.Add(time.Minute+time.Second)) {
		t.Fatal("expired")
	}
	m.MarkUnhealthy("worker", "stopped", now)
	if m.AllHealthy(now) {
		t.Fatal("unhealthy")
	}
	if m.Healthy("missing", now) {
		t.Fatal("missing healthy")
	}
}
