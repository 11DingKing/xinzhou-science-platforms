package registry

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type Item struct {
	ID        int64
	Name      string
	Kind      string
	Status    string
	Tags      []string
	CreatedAt time.Time
	Version   int64
}
type Registry struct {
	mu    sync.RWMutex
	items map[int64]Item
	next  int64
}

func New() *Registry { return &Registry{items: map[int64]Item{}, next: 1} }
func (r *Registry) Register(name, kind string, tags []string, now time.Time) (Item, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(kind) == "" {
		return Item{}, errors.New("registry name and kind required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.items {
		if item.Name == name && item.Kind == kind && item.Status == "active" {
			return Item{}, errors.New("active item already exists")
		}
	}
	item := Item{ID: r.next, Name: name, Kind: kind, Status: "active", Tags: append([]string(nil), tags...), CreatedAt: now.UTC(), Version: 1}
	r.items[item.ID] = item
	r.next++
	return item, nil
}
func (r *Registry) Retire(id, version int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	if !ok {
		return errors.New("item not found")
	}
	if item.Version != version {
		return errors.New("version conflict")
	}
	if item.Status != "active" {
		return errors.New("item not active")
	}
	item.Status = "retired"
	item.Version++
	r.items[id] = item
	return nil
}
func (r *Registry) List(kind, status string) []Item {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Item{}
	for _, item := range r.items {
		if (kind == "" || item.Kind == kind) && (status == "" || item.Status == status) {
			item.Tags = append([]string(nil), item.Tags...)
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Registry) Get(id int64) (Item, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	item.Tags = append([]string(nil), item.Tags...)
	return item, ok
}
