package etag

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

func Version(v int64) string { return `W/"` + strconv.FormatInt(v, 10) + `"` }
func Match(header string, v int64) bool {
	value := strings.TrimSpace(header)
	return value == Version(v) || value == `"`+strconv.FormatInt(v, 10)+`"`
}
func Body(payload []byte) string { sum := sha256.Sum256(payload); return hex.EncodeToString(sum[:]) }
func IfMatch(header string, v int64) error {
	if !Match(header, v) {
		return ErrMismatch
	}
	return nil
}

var ErrMismatch = errorString("etag mismatch")

type errorString string

func (e errorString) Error() string { return string(e) }
