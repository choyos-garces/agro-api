package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/choyos-garces/agro-api/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTemp(t, "{}\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("server.host: got %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8090 {
		t.Errorf("server.port: got %d, want 8090", cfg.Server.Port)
	}
	if cfg.Cookie.Name != "agro_session" {
		t.Errorf("cookie.name: got %q, want %q", cfg.Cookie.Name, "agro_session")
	}
	if cfg.Auth.CookieTTL.Duration != 24*time.Hour {
		t.Errorf("auth.cookieTTL: got %v, want 24h", cfg.Auth.CookieTTL.Duration)
	}
}

func TestLoad_FullConfig(t *testing.T) {
	yaml := `
server:
  host: "127.0.0.1"
  port: 9000

cors:
  allowedOrigins:
    - "https://example.com"
  allowedMethods:
    - GET
    - POST
  allowedHeaders:
    - Content-Type
  allowCredentials: true

cookie:
  name: "my_session"
  secure: true
  httpOnly: true
  sameSite: "Lax"
  path: "/api"

auth:
  cookieTTL: "12h"
  tokenSecret: "supersecret"
`
	path := writeTemp(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Addr() != "127.0.0.1:9000" {
		t.Errorf("Addr(): got %q, want %q", cfg.Server.Addr(), "127.0.0.1:9000")
	}
	if cfg.Auth.CookieTTL.Duration != 12*time.Hour {
		t.Errorf("cookieTTL: got %v, want 12h", cfg.Auth.CookieTTL.Duration)
	}
	if cfg.Cookie.SameSite != "Lax" {
		t.Errorf("sameSite: got %q, want Lax", cfg.Cookie.SameSite)
	}
	if cfg.Auth.TokenSecret != "supersecret" {
		t.Errorf("tokenSecret: got %q, want supersecret", cfg.Auth.TokenSecret)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	yaml := `
auth:
  cookieTTL: "notaduration"
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
}
