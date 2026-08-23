package etag

import (
	"testing"
)

func TestETagMatching(t *testing.T) {
	if !Match(`W/"7"`, 7) || !Match(`"7"`, 7) || Match(`W/"8"`, 7) {
		t.Fatal("match")
	}
	if IfMatch(`W/"8"`, 7) == nil {
		t.Fatal("mismatch accepted")
	}
	if Body([]byte("x")) == Body([]byte("y")) {
		t.Fatal("body hash")
	}
}
