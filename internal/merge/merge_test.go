package merge

import (
	"testing"
)

func TestPreferSources(t *testing.T) {
	got := Prefer([]Value{{Key: "sku", Source: "old", Priority: 1, UpdatedAt: 4, Payload: "a"}, {Key: "sku", Source: "new", Priority: 2, UpdatedAt: 1, Payload: "b"}, {Key: "region", Source: "x", Priority: 1, UpdatedAt: 1, Payload: "c"}})
	if len(got) != 2 || got[0].Key != "region" || got[1].Payload != "b" {
		t.Fatalf("merged=%+v", got)
	}
	if !Equal(Value{Key: " SKU ", Payload: "x"}, Value{Key: "sku", Payload: "x"}) {
		t.Fatal("equal")
	}
}
