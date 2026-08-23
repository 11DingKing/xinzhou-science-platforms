package clockwindow

import (
	"errors"
	"time"
)

type Window struct {
	Start, End time.Time
	Location   *time.Location
}

func New(start, end time.Time, loc *time.Location) (Window, error) {
	if loc == nil {
		loc = time.UTC
	}
	if end.Before(start) {
		return Window{}, errors.New("window reversed")
	}
	return Window{Start: start.In(loc), End: end.In(loc), Location: loc}, nil
}
func (w Window) Contains(value time.Time) bool {
	local := value.In(w.Location)
	return !local.Before(w.Start) && local.Before(w.End)
}
func (w Window) Duration() time.Duration    { return w.End.Sub(w.Start) }
func (w Window) Expired(now time.Time) bool { return !now.In(w.Location).Before(w.End) }
func (w Window) Remaining(now time.Time) time.Duration {
	if w.Expired(now) {
		return 0
	}
	return w.End.Sub(now.In(w.Location))
}
func Split(w Window, parts int) []Window {
	if parts < 1 || w.Duration() <= 0 {
		return nil
	}
	step := w.Duration() / time.Duration(parts)
	out := make([]Window, 0, parts)
	for i := 0; i < parts; i++ {
		start := w.Start.Add(step * time.Duration(i))
		end := start.Add(step)
		if i == parts-1 {
			end = w.End
		}
		out = append(out, Window{Start: start, End: end, Location: w.Location})
	}
	return out
}
