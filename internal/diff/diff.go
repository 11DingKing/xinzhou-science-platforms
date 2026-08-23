package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Change struct {
	Path   string
	Before any
	After  any
	Kind   string
}

func CompareMaps(before, after map[string]any) []Change {
	keys := map[string]bool{}
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}
	out := []Change{}
	for k := range keys {
		left, lok := before[k]
		right, rok := after[k]
		if !lok {
			out = append(out, Change{Path: k, After: right, Kind: "added"})
		} else if !rok {
			out = append(out, Change{Path: k, Before: left, Kind: "removed"})
		} else if !reflect.DeepEqual(left, right) {
			out = append(out, Change{Path: k, Before: left, After: right, Kind: "changed"})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
func Summary(changes []Change) string {
	parts := []string{}
	for _, c := range changes {
		parts = append(parts, fmt.Sprintf("%s:%s", c.Path, c.Kind))
	}
	return strings.Join(parts, ", ")
}
func HasKind(changes []Change, kind string) bool {
	for _, c := range changes {
		if c.Kind == kind {
			return true
		}
	}
	return false
}
func Only(changes []Change, allowed ...string) bool {
	set := map[string]bool{}
	for _, v := range allowed {
		set[v] = true
	}
	for _, c := range changes {
		if !set[c.Kind] {
			return false
		}
	}
	return true
}
