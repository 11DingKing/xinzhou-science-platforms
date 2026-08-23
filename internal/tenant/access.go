package tenant

import (
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
)

var ErrTenantMismatch = errors.New("tenant boundary violation")

type Scope struct {
	UserID     int64
	Role       domain.Role
	MerchantID int64
}

func (s Scope) CanReadMerchant(id int64) bool {
	return s.Role == domain.RoleReviewer || s.MerchantID == id
}
func (s Scope) CanWriteMerchant(id int64) bool {
	return s.Role == domain.RoleMerchant && s.MerchantID == id
}
func RequireRead(s Scope, id int64) error {
	if !s.CanReadMerchant(id) {
		return ErrTenantMismatch
	}
	return nil
}
func RequireWrite(s Scope, id int64) error {
	if !s.CanWriteMerchant(id) {
		return ErrTenantMismatch
	}
	return nil
}
func IsPlatform(r domain.Role) bool { return r == domain.RoleReviewer }
func IsMerchant(r domain.Role) bool { return r == domain.RoleMerchant }
