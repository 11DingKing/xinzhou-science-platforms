package command

import (
	"context"
	"testing"
	"time"
)

func TestCommandRegistry(t *testing.T) {
	r := New()
	called := false
	if err := r.Register("inspect", func(context.Context, any) error { called = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if r.Register("inspect", nil) == nil {
		t.Fatal("duplicate accepted")
	}
	if err := r.Execute(context.Background(), "inspect", nil); err != nil || !called {
		t.Fatal("execute")
	}
	if r.Execute(context.Background(), "missing", nil) == nil {
		t.Fatal("missing")
	}
	env := Envelope{ID: "1", Name: "inspect", CreatedAt: time.Now()}
	if !env.Valid() {
		t.Fatal("envelope")
	}
}
