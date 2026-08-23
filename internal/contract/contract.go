package contract

import (
	"errors"
	"strings"
	"time"
)

type Field struct {
	Name      string
	Required  bool
	MaxLength int
}
type Contract struct {
	Name       string
	Version    string
	Fields     []Field
	Deprecated bool
}
type Violation struct {
	Field  string
	Reason string
}

func (c Contract) Validate(payload map[string]string) []Violation {
	out := []Violation{}
	for _, f := range c.Fields {
		value, ok := payload[f.Name]
		if f.Required && (!ok || strings.TrimSpace(value) == "") {
			out = append(out, Violation{Field: f.Name, Reason: "required"})
			continue
		}
		if f.MaxLength > 0 && len(value) > f.MaxLength {
			out = append(out, Violation{Field: f.Name, Reason: "too long"})
		}
	}
	return out
}
func (c Contract) Valid() error {
	if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Version) == "" {
		return errors.New("contract name/version required")
	}
	seen := map[string]bool{}
	for _, f := range c.Fields {
		if f.Name == "" || seen[f.Name] {
			return errors.New("duplicate contract field")
		}
		seen[f.Name] = true
	}
	return nil
}
func Expired(created, now time.Time, ttl time.Duration) bool { return created.Add(ttl).Before(now) }
func RequiredNames(c Contract) []string {
	out := []string{}
	for _, f := range c.Fields {
		if f.Required {
			out = append(out, f.Name)
		}
	}
	return out
}
