package health

import (
	"context"
	"testing"
	"time"
)

func TestHealthStates(t *testing.T) {
	now := time.Now()
	if !Alive(now).Ready || Check(context.Background(), nil, now).Ready {
		t.Fatal("health")
	}
	if Degraded("maintenance", now).Message != "maintenance" {
		t.Fatal("degraded")
	}
}
