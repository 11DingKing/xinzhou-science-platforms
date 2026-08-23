package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/11DingKing/xinzhou-science-platforms/internal/auth"
	"github.com/11DingKing/xinzhou-science-platforms/internal/config"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func apiFixture(t *testing.T) (*API, context.Context, string) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file:"+filepath.Join(t.TempDir(), "http.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := repository.New(db)
	if _, err := r.CreateUser(ctx, "merchant", "pw", domain.RoleMerchant); err != nil {
		t.Fatal(err)
	}
	return New(config.Config{}, r, auth.NewService(r)), ctx, "merchant"
}
func reqJSON(t *testing.T, h http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
func TestHealthAndReady(t *testing.T) {
	api, _, _ := apiFixture(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		rec := reqJSON(t, api.Handler(), http.MethodGet, path, nil, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
	}
}
func TestAuthAndVersionHTTP(t *testing.T) {
	api, ctx, email := apiFixture(t)
	_ = ctx
	rec := reqJSON(t, api.Handler(), http.MethodPost, "/v1/auth/login", map[string]string{"email": email, "password": "pw"}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var login struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	if login.Token == "" {
		t.Fatal("empty token")
	}
	rec = reqJSON(t, api.Handler(), http.MethodPost, "/v1/versions", map[string]string{"sku": "x", "displayName": "Lamp", "channel": "online"}, login.Token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("version status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = reqJSON(t, api.Handler(), http.MethodPost, "/v1/auth/logout", nil, login.Token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status=%d", rec.Code)
	}
	rec = reqJSON(t, api.Handler(), http.MethodPost, "/v1/versions", map[string]string{"sku": "x2", "displayName": "Lamp", "channel": "online"}, login.Token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked status=%d", rec.Code)
	}
}
func TestUnauthorizedAndForbiddenHTTP(t *testing.T) {
	api, ctx, _ := apiFixture(t)
	_ = ctx
	rec := reqJSON(t, api.Handler(), http.MethodPost, "/v1/versions", map[string]string{}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized=%d", rec.Code)
	}
}
