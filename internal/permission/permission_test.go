package permission

import (
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"testing"
)

func TestPermissionMatrix(t *testing.T) {
	if Require(domain.RoleMerchant, ActionCreate, true) != nil || Require(domain.RoleMerchant, ActionReview, true) == nil {
		t.Fatal("merchant matrix")
	}
	if Require(domain.RoleReviewer, ActionExport, false) != nil {
		t.Fatal("reviewer export")
	}
	if len(Actions(domain.RoleMerchant, true)) != 2 || len(Actions(domain.RoleReviewer, false)) != 4 {
		t.Fatal("actions")
	}
}
