package diff

import (
	"strings"
	"testing"
)

func TestMapDiff(t *testing.T) {
	changes := CompareMaps(map[string]any{"a": 1, "b": "old"}, map[string]any{"a": 2, "c": true})
	if len(changes) != 3 || !HasKind(changes, "removed") || !Only(changes, "changed", "removed", "added") {
		t.Fatal("diff")
	}
	if !strings.Contains(Summary(changes), "a:changed") {
		t.Fatal("summary")
	}
}
