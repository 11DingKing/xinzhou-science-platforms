package timeline

import (
	"testing"
	"time"
)

func TestTimelineOrderingAndFiltering(t *testing.T) {
	base := time.Now()
	ts := Timeline{{At: base.Add(time.Hour), Kind: "b", ObjectID: 2}, {At: base, Kind: "a", ObjectID: 1}, {At: base, Kind: "a", ObjectID: 3}}
	ordered := ts.Ordered()
	if ordered[0].Kind != "a" || ordered[2].ObjectID != 2 {
		t.Fatalf("ordered=%+v", ordered)
	}
	if len(ts.Between(base, base.Add(30*time.Minute))) != 2 {
		t.Fatal("between")
	}
	if len(ts.ForObject(1)) != 1 || len(ts.Kinds()) != 2 {
		t.Fatal("filters")
	}
}
