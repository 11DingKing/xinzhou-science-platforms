package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"net/http"
)

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperrors.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperrors.ErrConflict), errors.Is(err, apperrors.ErrInvalidState):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
func writeJSONError(w http.ResponseWriter, r *http.Request, err error) {
	code := statusFor(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorBody{Code: http.StatusText(code), Message: err.Error(), RequestID: r.Header.Get("X-Request-ID")})
}
