package registry

import (
	"testing"
	"time"
)

func TestRegistryVersioning(t *testing.T) {
	r := New()
	item, err := r.Register("online-v1", "version", []string{"sku"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Register("online-v1", "version", nil, time.Now()); err == nil {
		t.Fatal("duplicate")
	}
	if err := r.Retire(item.ID, item.Version-1); err == nil {
		t.Fatal("stale retire")
	}
	if err := r.Retire(item.ID, item.Version); err != nil {
		t.Fatal(err)
	}
	if len(r.List("version", "retired")) != 1 {
		t.Fatal("retired")
	}
	got, _ := r.Get(item.ID)
	got.Tags[0] = "changed"
	again, _ := r.Get(item.ID)
	if again.Tags[0] == "changed" {
		t.Fatal("tag alias")
	}
}
