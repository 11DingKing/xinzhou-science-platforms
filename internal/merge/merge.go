package merge

import (
	"sort"
	"strings"
)

type Value struct {
	Key       string
	Source    string
	Priority  int
	UpdatedAt int64
	Payload   string
}

func Prefer(values []Value) []Value {
	groups := map[string][]Value{}
	for _, v := range values {
		groups[v.Key] = append(groups[v.Key], v)
	}
	out := []Value{}
	for _, items := range groups {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Priority == items[j].Priority {
				return items[i].UpdatedAt > items[j].UpdatedAt
			}
			return items[i].Priority > items[j].Priority
		})
		out = append(out, items[0])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
func Keys(values []Value) []string {
	out := []string{}
	for _, v := range Prefer(values) {
		out = append(out, v.Key)
	}
	return out
}
func NormalizeKey(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
func Equal(a, b Value) bool {
	return NormalizeKey(a.Key) == NormalizeKey(b.Key) && a.Payload == b.Payload
}
