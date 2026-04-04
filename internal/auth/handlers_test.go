package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/choyos-garces/agro-api/config"
	"github.com/choyos-garces/agro-api/internal/auth"
)

func testConfig() *config.Config {
	return &config.Config{
		Cookie: config.CookieConfig{
			Name:     "agro_session",
			Secure:   true,
			HTTPOnly: true,
			SameSite: "Strict",
			Path:     "/",
		},
		Auth: config.AuthConfig{
			CookieTTL: config.Duration{Duration: 24 * time.Hour},
		},
	}
}

func TestNew(t *testing.T) {
	cfg := testConfig()
	h := auth.New(cfg)
	if h == nil {
		t.Fatal("expected non-nil Handlers")
	}
}

// TestSetAndClearCookie verifies the cookie helpers through the exported
// helpers by checking that the response after logout has Max-Age=-1.
func TestLogoutClearsCookie(t *testing.T) {
	cfg := testConfig()
	h := auth.New(cfg)

	// Use the exported helper to simulate a login cookie.
	w := httptest.NewRecorder()
	h.SetSessionCookieForTest(w, "test-token")
	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected at least one cookie after login")
	}
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == cfg.Cookie.Name {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatalf("cookie %q not found", cfg.Cookie.Name)
	}
	if sessionCookie.Value != "test-token" {
		t.Errorf("cookie value: got %q, want %q", sessionCookie.Value, "test-token")
	}
	if !sessionCookie.HttpOnly {
		t.Error("expected HttpOnly=true")
	}
	if !sessionCookie.Secure {
		t.Error("expected Secure=true")
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite: got %v, want Strict", sessionCookie.SameSite)
	}

	// Now test clear.
	w2 := httptest.NewRecorder()
	h.ClearSessionCookieForTest(w2)
	resp2 := w2.Result()
	for _, c := range resp2.Cookies() {
		if c.Name == cfg.Cookie.Name {
			if c.MaxAge != -1 {
				t.Errorf("logout MaxAge: got %d, want -1", c.MaxAge)
			}
		}
	}
}
