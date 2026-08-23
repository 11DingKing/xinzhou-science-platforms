package query

import (
	"testing"
)

func TestQueryHelpers(t *testing.T) {
	p := Page{Number: 0, Size: 200}
	if p.Normalize().Size != 100 {
		t.Fatal("page")
	}
	p = Page{Number: 2, Size: 10}
	if p.Offset() != 10 {
		t.Fatal("offset")
	}
	s := Sort{Field: "created", Descending: true}
	if clause, err := s.Clause(map[string]string{"created": "created_at"}); err != nil || clause != "created_at DESC" {
		t.Fatal("sort")
	}
	s = Sort{Field: "bad"}
	if _, err := s.Clause(map[string]string{"created": "created_at"}); err == nil {
		t.Fatal("bad sort")
	}
}
