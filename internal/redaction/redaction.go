package redaction

import (
	"regexp"
	"strings"
)

var email = regexp.MustCompile(`(?i)([a-z0-9._%+\-]+)@([a-z0-9.\-]+\.[a-z]{2,})`)
var token = regexp.MustCompile(`(?i)(bearer\s+|token[=:]\s*)([a-z0-9._\-]+)`)

func Text(value string) string {
	value = email.ReplaceAllString(value, "$1***@$2")
	return token.ReplaceAllString(value, "$1***")
}
func Fields(values map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range values {
		if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "secret") {
			out[k] = "***"
		} else {
			out[k] = Text(v)
		}
	}
	return out
}
