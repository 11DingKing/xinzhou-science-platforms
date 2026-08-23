package response

import (
	"encoding/json"
	"net/http"
	"time"
)

type Meta struct {
	RequestID   string    `json:"request_id"`
	GeneratedAt time.Time `json:"generated_at"`
}
type Envelope[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

func Write[T any](w http.ResponseWriter, status int, value T, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope[T]{Data: value, Meta: Meta{RequestID: requestID, GeneratedAt: time.Now().UTC()}})
}
func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }
func Accepted(w http.ResponseWriter, value any, requestID string) {
	Write(w, http.StatusAccepted, value, requestID)
}
