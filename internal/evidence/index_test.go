package evidence

import (
	"testing"
	"time"
)

func TestEvidenceIndex(t *testing.T) {
	i := NewIndex()
	item, err := i.Add(1, "photo", "a.jpg", []byte("bytes"), time.Now())
	if err != nil || item.Hash == "" {
		t.Fatal(err)
	}
	if i.Pending(1) != 1 || i.AllArchived(1) {
		t.Fatal("pending")
	}
	if err := i.Archive(item.ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if !i.AllArchived(1) {
		t.Fatal("archive")
	}
	if err := i.Archive(item.ID, time.Now()); err == nil {
		t.Fatal("double archive")
	}
}
