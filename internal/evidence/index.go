package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

type Item struct {
	ID         int64
	CaseID     int64
	Kind       string
	ObjectKey  string
	Hash       string
	Size       int64
	CreatedAt  time.Time
	ArchivedAt *time.Time
}
type Index struct {
	items map[int64]Item
	next  int64
}

func NewIndex() *Index { return &Index{items: map[int64]Item{}, next: 1} }
func (i *Index) Add(caseID int64, kind, key string, data []byte, now time.Time) (Item, error) {
	if caseID < 1 || strings.TrimSpace(kind) == "" || strings.TrimSpace(key) == "" || len(data) == 0 {
		return Item{}, errors.New("invalid evidence")
	}
	sum := sha256.Sum256(data)
	item := Item{ID: i.next, CaseID: caseID, Kind: kind, ObjectKey: key, Hash: hex.EncodeToString(sum[:]), Size: int64(len(data)), CreatedAt: now.UTC()}
	i.items[item.ID] = item
	i.next++
	return item, nil
}
func (i *Index) Archive(id int64, now time.Time) error {
	item, ok := i.items[id]
	if !ok {
		return errors.New("evidence not found")
	}
	if item.ArchivedAt != nil {
		return errors.New("evidence already archived")
	}
	t := now.UTC()
	item.ArchivedAt = &t
	i.items[id] = item
	return nil
}
func (i *Index) ForCase(caseID int64) []Item {
	out := []Item{}
	for _, item := range i.items {
		if item.CaseID == caseID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].ID < out[b].ID })
	return out
}
func (i *Index) Pending(caseID int64) int {
	n := 0
	for _, item := range i.ForCase(caseID) {
		if item.ArchivedAt == nil {
			n++
		}
	}
	return n
}
func (i *Index) AllArchived(caseID int64) bool {
	return len(i.ForCase(caseID)) > 0 && i.Pending(caseID) == 0
}
