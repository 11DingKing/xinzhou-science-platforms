package auth

import (
	"context"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"path/filepath"
	"testing"
)

func TestLoginLogoutAndRole(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := repository.New(db)
	_, err = r.CreateUser(ctx, "merchant", "secret", domain.RoleMerchant)
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(r)
	token, u, err := s.Login(ctx, "merchant", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || u.Role != domain.RoleMerchant {
		t.Fatalf("token=%q user=%+v", token, u)
	}
	if _, _, err := s.Login(ctx, "merchant", "wrong"); !errors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("wrong password err=%v", err)
	}
	if err := s.Logout(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(ctx, token); !errors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("logout auth err=%v", err)
	}
}
