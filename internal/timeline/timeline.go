package timeline

import (
	"sort"
	"time"
)

type Event struct {
	At       time.Time
	Kind     string
	Actor    int64
	ObjectID int64
	Message  string
}
type Timeline []Event

func (t Timeline) Add(e Event) Timeline { return append(append(Timeline(nil), t...), e) }
func (t Timeline) Ordered() Timeline {
	out := append(Timeline(nil), t...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].At.Equal(out[j].At) {
			return out[i].ObjectID < out[j].ObjectID
		}
		return out[i].At.Before(out[j].At)
	})
	return out
}
func (t Timeline) Between(from, to time.Time) Timeline {
	out := Timeline{}
	for _, e := range t {
		if !e.At.Before(from) && e.At.Before(to) {
			out = append(out, e)
		}
	}
	return out
}
func (t Timeline) ForObject(id int64) Timeline {
	out := Timeline{}
	for _, e := range t {
		if e.ObjectID == id {
			out = append(out, e)
		}
	}
	return out
}
func (t Timeline) Kinds() []string {
	out := []string{}
	seen := map[string]bool{}
	for _, e := range t {
		if !seen[e.Kind] {
			seen[e.Kind] = true
			out = append(out, e.Kind)
		}
	}
	return out
}
