// Package auth provides HTTP handlers for session-based authentication
// built on top of PocketBase's built-in user management.
//
// Endpoints:
//
//	POST /api/auth/login   – exchange email/password for a secure session cookie
//	GET  /api/auth/verify  – validate the current session cookie
//	POST /api/auth/logout  – destroy the session cookie
package auth

import (
	"net/http"
	"time"

	"github.com/choyos-garces/agro-api/config"
	"github.com/pocketbase/pocketbase/core"
)

// Handlers groups all auth-related route handlers together with their shared
// dependencies so they can be registered on the PocketBase router.
type Handlers struct {
	cfg *config.Config
}

// New creates a new Handlers instance bound to the given Config.
func New(cfg *config.Config) *Handlers {
	return &Handlers{cfg: cfg}
}

// RegisterRoutes attaches the auth endpoints to the PocketBase app router.
func (h *Handlers) RegisterRoutes(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/auth/login", h.Login)
		se.Router.GET("/api/auth/verify", h.Verify)
		se.Router.POST("/api/auth/logout", h.Logout)
		return se.Next()
	})
}

// loginRequest is the expected JSON body for the login endpoint.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user with their email and password, then sets a
// secure HTTP-only session cookie containing the PocketBase auth token.
//
//	POST /api/auth/login
//	Content-Type: application/json
//	{"email":"user@example.com","password":"secret"}
func (h *Handlers) Login(e *core.RequestEvent) error {
	var req loginRequest
	if err := e.BindBody(&req); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if req.Email == "" || req.Password == "" {
		return e.BadRequestError("email and password are required", nil)
	}

	// Authenticate against the PocketBase _superusers or users collection.
	// AuthWithPassword tries _superusers first; for regular users the collection
	// must be specified.  We attempt "users" (the default PocketBase collection).
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

	h.setSessionCookie(e.Response, token)

	return e.JSON(http.StatusOK, map[string]any{
		"message": "login successful",
		"id":      record.Id,
		"email":   record.Email(),
	})
}

// Verify checks whether the current session cookie holds a valid PocketBase
// auth token.  Returns 200 when valid, 401 when missing or invalid.
//
//	GET /api/auth/verify
func (h *Handlers) Verify(e *core.RequestEvent) error {
	cookie, err := e.Request.Cookie(h.cfg.Cookie.Name)
	if err != nil {
		return e.UnauthorizedError("no session cookie", nil)
	}

	record, err := e.App.FindAuthRecordByToken(cookie.Value, core.TokenTypeAuth)
	if err != nil {
		return e.UnauthorizedError("invalid or expired session", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"valid": true,
		"id":    record.Id,
		"email": record.Email(),
	})
}

// Logout destroys the session by overwriting the cookie with an expired one.
//
//	POST /api/auth/logout
func (h *Handlers) Logout(e *core.RequestEvent) error {
	h.clearSessionCookie(e.Response)
	return e.JSON(http.StatusOK, map[string]any{"message": "logged out"})
}

// --- helpers -----------------------------------------------------------------

// SetSessionCookieForTest is exported only for use in package tests.
// Production code should call setSessionCookie instead.
func (h *Handlers) SetSessionCookieForTest(w http.ResponseWriter, token string) {
	h.setSessionCookie(w, token)
}

// ClearSessionCookieForTest is exported only for use in package tests.
// Production code should call clearSessionCookie instead.
func (h *Handlers) ClearSessionCookieForTest(w http.ResponseWriter) {
	h.clearSessionCookie(w)
}

func (h *Handlers) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.Cookie.Name,
		Value:    token,
		Path:     h.cfg.Cookie.Path,
		MaxAge:   int(h.cfg.Auth.CookieTTL.Duration / time.Second),
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: h.cfg.Cookie.HTTPOnly,
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

func (h *Handlers) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.Cookie.Name,
		Value:    "",
		Path:     h.cfg.Cookie.Path,
		MaxAge:   -1,
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: h.cfg.Cookie.HTTPOnly,
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

// parseSameSite converts the string representation from the config file into
// the http.SameSite enum value.
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
