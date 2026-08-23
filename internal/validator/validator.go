package validator

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var ErrRequired = errors.New("required value")
var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{2,31}$`)

func Email(v string) bool            { _, err := mail.ParseAddress(strings.TrimSpace(v)); return err == nil }
func BatchCode(v string) bool        { return codePattern.MatchString(strings.TrimSpace(v)) }
func Region(v string) bool           { return len(strings.TrimSpace(v)) >= 2 && len(strings.TrimSpace(v)) <= 64 }
func Window(from, to time.Time) bool { return !from.IsZero() && !to.Before(from) }
func Required(values ...string) error {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return ErrRequired
		}
	}
	return nil
}
func Unique(values []string) bool {
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}
