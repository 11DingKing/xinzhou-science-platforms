package tenant

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"testing"
)

func TestMerchantScope(t *testing.T) {
	s := Scope{Role: domain.RoleMerchant, MerchantID: 7}
	if RequireRead(s, 7) != nil || RequireWrite(s, 7) != nil {
		t.Fatal("owner denied")
	}
	if RequireRead(s, 8) == nil || RequireWrite(s, 8) == nil {
		t.Fatal("cross merchant access")
	}
}
func TestReviewerScope(t *testing.T) {
	s := Scope{Role: domain.RoleReviewer}
	if RequireRead(s, 99) != nil {
		t.Fatal("reviewer read denied")
	}
	if RequireWrite(s, 99) == nil {
		t.Fatal("reviewer write accepted")
	}
	if !IsPlatform(s.Role) || IsMerchant(s.Role) {
		t.Fatal("role helper")
	}
}
