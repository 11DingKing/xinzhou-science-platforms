package tenant

import (
	"testing"
)

func TestTenantFilter(t *testing.T) {
	f := Filter{MerchantID: 4}
	sql, args := f.SQL("p")
	if sql != "p.merchant_id=?" || len(args) != 1 {
		t.Fatal("filter")
	}
	if f.Validate() != nil {
		t.Fatal("valid")
	}
	empty := Filter{}
	if empty.Validate() == nil {
		t.Fatal("empty")
	}
	if NormalizeAlias(" P ") != "p" || !SameTenant(f, Filter{MerchantID: 4}) {
		t.Fatal("helpers")
	}
}
