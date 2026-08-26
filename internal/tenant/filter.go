package tenant

import (
	"fmt"
	"strings"
)

type Filter struct {
	MerchantID int64
	Reviewer   bool
	Alias      string
}

func (f Filter) SQL(alias string) (string, []any) {
	if f.Reviewer {
		return "1=1", nil
	}
	if alias == "" {
		alias = f.Alias
	}
	if alias == "" {
		alias = "p"
	}
	return fmt.Sprintf("%s.merchant_id=?", alias), []any{f.MerchantID}
}
func (f Filter) Validate() error {
	if f.Reviewer {
		return nil
	}
	if f.MerchantID < 1 {
		return ErrTenantMismatch
	}
	return nil
}
func NormalizeAlias(v string) string { return strings.Trim(strings.ToLower(v), " ") }
func SameTenant(a, b Filter) bool {
	return a.Reviewer == b.Reviewer && (a.Reviewer || a.MerchantID == b.MerchantID)
}
