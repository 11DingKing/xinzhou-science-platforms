package pagination

import (
	"net/url"
	"testing"
)

func TestParseClampsAndNormalizes(t *testing.T) {
	r := Parse(url.Values{"page": []string{"0"}, "page_size": []string{"500"}, "direction": []string{"DESC"}, "search": []string{" lamp "}})
	if r.Page != 1 || r.PageSize != 100 || r.Direction != "desc" || r.Search != "lamp" {
		t.Fatalf("request=%+v", r)
	}
	if r.Offset() != 0 {
		t.Fatal("offset")
	}
}
func TestQueryBuild(t *testing.T) {
	r := Parse(url.Values{"region": []string{"county"}, "search": []string{"lamp"}})
	where, args := BuildWhere(r)
	if len(args) != 3 || where == "" {
		t.Fatalf("where=%s args=%v", where, args)
	}
	if r.Validate(map[string]bool{"created_at": true}) != nil {
		t.Fatal("empty sort invalid")
	}
	r.Sort = "bad"
	if r.Validate(map[string]bool{"created_at": true}) == nil {
		t.Fatal("bad sort accepted")
	}
}
func TestResultHelpers(t *testing.T) {
	if ClampTotal(41, 20) != 3 || ClampTotal(-1, 20) != 0 {
		t.Fatal("page count")
	}
	got := Map([]int{1, 2, 3}, func(v int) int { return v * v })
	if len(got) != 3 || got[2] != 9 {
		t.Fatal("map")
	}
	if len(Empty[int]().Items) != 0 {
		t.Fatal("empty result")
	}
}
