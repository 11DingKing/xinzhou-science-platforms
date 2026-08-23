package queue

import (
	"context"
	"testing"
	"time"
)

func TestQueueOrderAndClose(t *testing.T) {
	q := New[string]()
	if err := q.Push("one"); err != nil {
		t.Fatal(err)
	}
	if err := q.Push("two"); err != nil {
		t.Fatal(err)
	}
	a, _ := q.Pop(context.Background())
	b, _ := q.Pop(context.Background())
	if a.Value != "one" || b.Value != "two" {
		t.Fatal("order")
	}
	q.Close()
	if err := q.Push("three"); err != ErrClosed {
		t.Fatal("push closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, err := q.Pop(ctx); err != ErrClosed {
		t.Fatalf("pop closed=%v", err)
	}
}
