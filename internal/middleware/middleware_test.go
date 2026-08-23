package middleware

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/metrics"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverAndObserve(t *testing.T) {
	counter := metrics.NewCounter()
	hist := metrics.NewHistogram()
	h := Observe(counter, hist)(Recover(slog.Default())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != 500 || counter.Get("GET /x") != 1 {
		t.Fatalf("status=%d count=%d", rec.Code, counter.Get("GET /x"))
	}
}
func TestRequireMethod(t *testing.T) {
	h := RequireMethod(http.MethodPost)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(201) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 405 || rec.Header().Get("Allow") != "POST" {
		t.Fatal("method")
	}
}
