package permission

import (
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
)

var ErrDenied = errors.New("permission denied")

type Action string

const (
	ActionRead   Action = "read"
	ActionCreate Action = "create"
	ActionReview Action = "review"
	ActionClose  Action = "close"
	ActionExport Action = "export"
)

func Allowed(role domain.Role, action Action, owner bool) bool {
	if role == domain.RoleReviewer {
		return action == ActionRead || action == ActionReview || action == ActionClose || action == ActionExport
	}
	if role == domain.RoleMerchant {
		return owner && (action == ActionRead || action == ActionCreate)
	}
	return false
}
func Require(role domain.Role, action Action, owner bool) error {
	if !Allowed(role, action, owner) {
		return ErrDenied
	}
	return nil
}
func Actions(role domain.Role, owner bool) []Action {
	all := []Action{ActionRead, ActionCreate, ActionReview, ActionClose, ActionExport}
	out := []Action{}
	for _, a := range all {
		if Allowed(role, a, owner) {
			out = append(out, a)
		}
	}
	return out
}
