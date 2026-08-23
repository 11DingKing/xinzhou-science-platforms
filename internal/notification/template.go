package notification

import (
	"fmt"
	"strings"
)

type Template struct {
	Kind     string
	Subject  string
	Body     string
	Required []string
}

func Render(t Template, values map[string]string) (string, error) {
	out := t.Body
	for _, key := range t.Required {
		value, ok := values[key]
		if !ok || strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("missing template value %s", key)
		}
		out = strings.ReplaceAll(out, "{{"+key+"}}", value)
	}
	return out, nil
}
func DefaultTemplates() map[string]Template {
	return map[string]Template{"complaint": {Kind: "complaint", Subject: "Quality complaint", Body: "Batch {{batch}} requires review in {{region}}.", Required: []string{"batch", "region"}}, "remediation": {Kind: "remediation", Subject: "Remediation due", Body: "Action {{action}} is due.", Required: []string{"action"}}}
}
func IsKnown(kind string) bool { _, ok := DefaultTemplates()[kind]; return ok }
