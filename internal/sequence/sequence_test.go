package sequence

import (
	"testing"
)

func TestSequenceGenerator(t *testing.T) {
	g := New("B", 3)
	if g.Next() != "B-3" || g.Next() != "B-4" {
		t.Fatal("sequence")
	}
	first, last, err := g.Reserve(2)
	if err != nil || first != "B-5" || last != "B-6" {
		t.Fatal("reserve")
	}
	if g.Current() != 7 {
		t.Fatal("current")
	}
	if _, _, err := g.Reserve(0); err == nil {
		t.Fatal("zero reserve")
	}
}
