package serialization

import (
	"testing"
	"time"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	data, err := Encode("inspection.submitted", "req-1", map[string]any{"id": 4}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	e, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]int
	if err := DecodePayload(e, &payload); err != nil || payload["id"] != 4 {
		t.Fatalf("payload=%v err=%v", payload, err)
	}
	copy := ClonePayload(e.Payload)
	copy[0] = 'x'
	if string(copy) == string(e.Payload) {
		t.Fatal("payload not cloned")
	}
}
