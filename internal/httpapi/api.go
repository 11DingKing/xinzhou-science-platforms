package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/apperrors"
	"github.com/11DingKing/xinzhou-science-platforms/internal/auth"
	"github.com/11DingKing/xinzhou-science-platforms/internal/config"
	"github.com/11DingKing/xinzhou-science-platforms/internal/domain"
	"github.com/11DingKing/xinzhou-science-platforms/internal/platform"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"net/http"
	"strings"
	"time"
)

type API struct {
	cfg       config.Config
	repos     *repository.Repos
	auth      *auth.Service
	platforms *platform.Service
}

func New(c config.Config, r *repository.Repos, a *auth.Service) *API {
	return &API{cfg: c, repos: r, auth: a}
}

func (a *API) SetPlatformService(s *platform.Service) { a.platforms = s }
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	mux.HandleFunc("POST /v1/auth/login", a.login)
	mux.HandleFunc("POST /v1/auth/logout", a.logout)
	mux.HandleFunc("POST /v1/versions", a.createVersion)
	mux.HandleFunc("POST /v1/versions/publish", a.publishVersion)
	mux.HandleFunc("POST /v1/batches", a.createBatch)
	mux.HandleFunc("POST /v1/platforms", a.createPlatform)
	mux.HandleFunc("POST /v1/platforms/submit", a.submitPlatform)
	mux.HandleFunc("POST /v1/platforms/review", a.reviewPlatform)
	mux.HandleFunc("POST /v1/platforms/activate", a.activatePlatform)
	mux.HandleFunc("POST /v1/platforms/milestones", a.addPlatformMilestone)
	mux.HandleFunc("POST /v1/platforms/funding", a.fundPlatform)
	mux.HandleFunc("POST /v1/platforms/reports", a.submitPlatformReport)
	return requestID(mux)
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r)
	})
}
func (a *API) user(r *http.Request) (domain.User, error) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return domain.User{}, apperrors.ErrUnauthorized
	}
	return a.auth.Authenticate(r.Context(), strings.TrimPrefix(h, "Bearer "))
}
func writeErr(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if errors.Is(err, apperrors.ErrUnauthorized) {
		code = http.StatusUnauthorized
	}
	if errors.Is(err, apperrors.ErrForbidden) {
		code = http.StatusForbidden
	}
	if errors.Is(err, apperrors.ErrConflict) || errors.Is(err, apperrors.ErrInvalidState) {
		code = http.StatusConflict
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, apperrors.ErrUnauthorized)
		return
	}
	token, u, err := a.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"token": token, "role": u.Role})
}
func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		writeErr(w, apperrors.ErrUnauthorized)
		return
	}
	if err := a.auth.Logout(r.Context(), strings.TrimPrefix(h, "Bearer ")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (a *API) createVersion(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if u.Role != domain.RoleMerchant {
		writeErr(w, apperrors.ErrForbidden)
		return
	}
	var v domain.ProductVersion
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	out, err := a.repos.CreateVersion(r.Context(), u, v)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out)
}
func (a *API) publishVersion(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct{ ID, Version int64 }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	if err := a.repos.PublishVersion(r.Context(), u, in.ID, in.Version); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (a *API) createBatch(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var b domain.Batch
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	out, err := a.repos.CreateBatch(r.Context(), u, b)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out)
}

func (a *API) platformActor(r *http.Request, roles ...domain.Role) (domain.User, error) {
	u, err := a.user(r)
	if err != nil {
		return u, err
	}
	for _, role := range roles {
		if u.Role == role {
			return u, nil
		}
	}
	return u, apperrors.ErrForbidden
}

func (a *API) createPlatform(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformLead, domain.RolePlatformAdmin)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct {
		Name, City, FocusArea string
		PlannedYear           int
		BudgetCents           int64
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	p, err := a.platforms.Create(r.Context(), u.ID, in.Name, in.City, in.FocusArea, in.PlannedYear, in.BudgetCents, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

func (a *API) submitPlatform(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformLead, domain.RolePlatformAdmin)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct{ ID, Version int64 }
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	if err := a.platforms.Submit(r.Context(), u.ID, in.ID, in.Version); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) reviewPlatform(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformAdmin, domain.RoleReviewer)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct {
		ID, Version int64
		Approve     bool
		Notes       string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	if err := a.platforms.StartReview(r.Context(), u.ID, in.ID, in.Version); err != nil {
		writeErr(w, err)
		return
	}
	if err := a.platforms.Decide(r.Context(), u.ID, in.ID, in.Version+1, in.Approve, in.Notes); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) activatePlatform(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformAdmin)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct{ ID, Version int64 }
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	if err := a.platforms.Activate(r.Context(), u.ID, in.ID, in.Version); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) addPlatformMilestone(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformLead, domain.RolePlatformAdmin)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct {
		PlatformID   int64
		Title, DueAt string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	due, err := time.Parse(time.RFC3339, in.DueAt)
	if err != nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	m, err := a.platforms.AddMilestone(r.Context(), u.ID, in.PlatformID, in.Title, due)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(m)
}

func (a *API) fundPlatform(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformAdmin)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct {
		PlatformID, AmountCents int64
		Tranche                 int
		IdempotencyKey          string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	if in.IdempotencyKey == "" {
		in.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	if err := a.platforms.RecordFunding(r.Context(), u.ID, in.PlatformID, in.AmountCents, in.Tranche, in.IdempotencyKey); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) submitPlatformReport(w http.ResponseWriter, r *http.Request) {
	u, err := a.platformActor(r, domain.RolePlatformLead, domain.RolePlatformAdmin)
	if err != nil {
		writeErr(w, err)
		return
	}
	var in struct {
		PlatformID int64
		Year       int
		Summary    string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || a.platforms == nil {
		writeErr(w, apperrors.ErrInvalidState)
		return
	}
	if err := a.platforms.SubmitReport(r.Context(), u.ID, in.PlatformID, in.Year, in.Summary); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
