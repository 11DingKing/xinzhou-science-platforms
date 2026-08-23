package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	Write(rec, http.StatusOK, map[string]string{"status": "ok"}, "r1")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "r1") {
		t.Fatal("response")
	}
	rec = httptest.NewRecorder()
	NoContent(rec)
	if rec.Code != 204 {
		t.Fatal("no content")
	}
}
