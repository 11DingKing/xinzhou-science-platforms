package eventbus

import (
	"context"
	"errors"
	"testing"
)

func TestBusPublishesInOrder(t *testing.T) {
	b := New()
	seen := []string{}
	b.Subscribe("x", func(_ context.Context, e Event) error { seen = append(seen, e.Payload.(string)); return nil })
	b.Subscribe("x", func(_ context.Context, e Event) error { seen = append(seen, "second"); return nil })
	if err := b.Publish(context.Background(), Event{Topic: "x", Payload: "first"}); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 || seen[0] != "first" || len(b.Topics()) != 1 {
		t.Fatalf("seen=%v", seen)
	}
}
func TestBusStopsOnHandlerError(t *testing.T) {
	b := New()
	called := 0
	b.Subscribe("x", func(context.Context, Event) error { called++; return errors.New("stop") })
	b.Subscribe("x", func(context.Context, Event) error { called++; return nil })
	if b.Publish(context.Background(), Event{Topic: "x"}) == nil || called != 1 {
		t.Fatalf("called=%d", called)
	}
}
