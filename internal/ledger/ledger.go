package ledger

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Entry struct {
	ID        int64
	Kind      string
	ObjectID  int64
	ActorID   int64
	Result    string
	At        time.Time
	RequestID string
}
type Ledger struct {
	mu      sync.RWMutex
	entries []Entry
	next    int64
}

func New() *Ledger { return &Ledger{next: 1} }
func (l *Ledger) Append(kind string, objectID, actorID int64, result, requestID string, at time.Time) (Entry, error) {
	if kind == "" || objectID < 1 || actorID < 1 || requestID == "" {
		return Entry{}, errors.New("invalid ledger entry")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Entry{ID: l.next, Kind: kind, ObjectID: objectID, ActorID: actorID, Result: result, RequestID: requestID, At: at.UTC()}
	l.entries = append(l.entries, e)
	l.next++
	return e, nil
}
func (l *Ledger) List(kind string, objectID int64) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := []Entry{}
	for _, e := range l.entries {
		if e.Kind == kind && e.ObjectID == objectID {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (l *Ledger) Since(at time.Time) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := []Entry{}
	for _, e := range l.entries {
		if !e.At.Before(at) {
			out = append(out, e)
		}
	}
	return out
}
func (l *Ledger) Count() int { l.mu.RLock(); defer l.mu.RUnlock(); return len(l.entries) }
