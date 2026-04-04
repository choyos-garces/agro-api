package auth

import (
	"net/http"
	"time"

	"github.com/choyos-garces/agro-api/config"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

type Handlers struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Handlers {
	return &Handlers{cfg: cfg}
}

func (h *Handlers) RegisterRoutes(r *router.Router[*core.RequestEvent]) {
	r.POST("/api/auth/login", h.Login)
	r.GET("/api/auth/verify", h.Verify)
	r.POST("/api/auth/logout", h.Logout)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type successResponse struct {
	ID     string   `json:"id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	Active bool     `json:"active"`
}

func (h *Handlers) Login(e *core.RequestEvent) error {
	var req loginRequest
	if err := e.BindBody(&req); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if req.Email == "" || req.Password == "" {
		return e.BadRequestError("email and password are required", nil)
	}

	record, err := e.App.FindAuthRecordByEmail("users", req.Email)
	if err != nil {
		return e.UnauthorizedError("invalid credentials", nil)
	}

	if !record.ValidatePassword(req.Password) {
		return e.UnauthorizedError("invalid credentials", nil)
	}

	token, err := record.NewAuthToken()
	if err != nil {
		return e.InternalServerError("could not create session token", err)
	}

	// Use PocketBase's native event context to set the cookie
	h.setSessionCookie(e, token)

	return e.JSON(http.StatusOK, successResponse{
		ID:     record.Id,
		Email:  record.Email(),
		Name:   record.GetString("name"),
		Roles:  []string{}, // placeholder for role management
		Active: true,
	})
}

func (h *Handlers) Verify(e *core.RequestEvent) error {
	cookie, err := e.Request.Cookie(h.cfg.Cookie.Name)
	if err != nil {
		return e.UnauthorizedError("no session cookie", nil)
	}

	record, err := e.App.FindAuthRecordByToken(cookie.Value, core.TokenTypeAuth)
	if err != nil {
		return e.UnauthorizedError("invalid or expired session", nil)
	}

	return e.JSON(http.StatusOK, successResponse{
		ID:     record.Id,
		Email:  record.Email(),
		Name:   record.GetString("name"),
		Roles:  []string{}, // placeholder for role management
		Active: true,
	})
}

func (h *Handlers) Logout(e *core.RequestEvent) error {
	h.clearSessionCookie(e)
	return e.JSON(http.StatusOK, map[string]any{"message": "logged out"})
}

func (h *Handlers) setSessionCookie(e *core.RequestEvent, token string) {
	e.SetCookie(&http.Cookie{
		Name:     h.cfg.Cookie.Name,
		Value:    token,
		Path:     h.cfg.Cookie.Path,
		MaxAge:   int(h.cfg.Auth.CookieTTL.Duration / time.Second),
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: h.cfg.Cookie.HTTPOnly,
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

func (h *Handlers) clearSessionCookie(e *core.RequestEvent) {
	e.SetCookie(&http.Cookie{
		Name:     h.cfg.Cookie.Name,
		Value:    "",
		Path:     h.cfg.Cookie.Path,
		MaxAge:   -1,
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: h.cfg.Cookie.HTTPOnly,
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
