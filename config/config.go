package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Server ServerConfig `yaml:"server"`
	CORS   CORSConfig   `yaml:"cors"`
	Cookie CookieConfig `yaml:"cookie"`
	Auth   AuthConfig   `yaml:"auth"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Addr returns the formatted "host:port" address.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// CORSConfig holds Cross-Origin Resource Sharing settings.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowedOrigins"`
	AllowedMethods   []string `yaml:"allowedMethods"`
	AllowedHeaders   []string `yaml:"allowedHeaders"`
	AllowCredentials bool     `yaml:"allowCredentials"`
}

// CookieConfig holds session cookie settings.
type CookieConfig struct {
	Name     string `yaml:"name"`
	Secure   bool   `yaml:"secure"`
	HTTPOnly bool   `yaml:"httpOnly"`
	SameSite string `yaml:"sameSite"`
	Path     string `yaml:"path"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	CookieTTL   Duration `yaml:"cookieTTL"`
	TokenSecret string   `yaml:"tokenSecret"`
}

// Duration is a wrapper around time.Duration to support YAML unmarshalling
// of human-readable duration strings (e.g. "24h", "30m").
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	dur, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = dur
	return nil
}

// MarshalYAML serialises the Duration as a string.
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// Load reads the YAML configuration file at the given path and returns a
// populated Config.  Missing optional fields fall back to their zero values;
// callers may apply defaults after loading.
func Load(path string) (*Config, error) {
	f, err := os.Open(path) //nolint:gosec // path is controlled by the caller
	if err != nil {
		return nil, fmt.Errorf("open config %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

// applyDefaults fills in sensible defaults for any zero-value fields.
func applyDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8090
	}
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"*"}
	}
	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(cfg.CORS.AllowedHeaders) == 0 {
		cfg.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}
	if cfg.Cookie.Name == "" {
		cfg.Cookie.Name = "agro_session"
	}
	if cfg.Cookie.Path == "" {
		cfg.Cookie.Path = "/"
	}
	if cfg.Cookie.SameSite == "" {
		cfg.Cookie.SameSite = "Strict"
	}
	if cfg.Auth.CookieTTL.Duration == 0 {
		cfg.Auth.CookieTTL.Duration = 24 * time.Hour
	}
}
