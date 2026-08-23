package middleware

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/metrics"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if value := recover(); value != nil {
					log.Error("panic recovered", "panic", value, "stack", string(debug.Stack()))
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
func Observe(counter *metrics.Counter, hist *metrics.Histogram) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			counter.Inc(r.Method + " " + r.URL.Path)
			hist.Observe(r.Method+" "+r.URL.Path, time.Since(start))
		})
	}
}
func RequireMethod(method string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				w.Header().Set("Allow", method)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
