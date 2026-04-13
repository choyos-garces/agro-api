package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Server ServerConfig `yaml:"server"`
	CORS   CORSConfig   `yaml:"cors"`
	Dev    bool         `yaml:"dev"`
	Google GoogleConfig `yaml:"google"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type GoogleConfig struct {
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
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

// Load reads the base YAML configuration file.
// If a ".local" version of the file exists (e.g., config.local.yaml),
// it loads that file as well, overriding the base values.
func Load(basePath string) (*Config, error) {
	var cfg Config

	// 1. Load the base config (e.g., config.yaml)
	// We expect this file to exist. If it doesn't, we return an error.
	if err := decodeFile(basePath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load base config %q: %w", basePath, err)
	}

	// 2. Generate the local file path (config.yaml -> config.local.yaml)
	ext := filepath.Ext(basePath)                       // gets ".yaml"
	baseWithoutExt := strings.TrimSuffix(basePath, ext) // gets "config"
	localPath := baseWithoutExt + ".local" + ext        // builds "config.local.yaml"

	// 3. Check if the local override file exists
	if _, err := os.Stat(localPath); err == nil {
		// The file exists! Decode it over the SAME struct instance.
		// It will automatically replace only the values defined in the local file.
		if err := decodeFile(localPath, &cfg); err != nil {
			return nil, fmt.Errorf("failed to load local override %q: %w", localPath, err)
		}
	}

	// 4. Fill in any gaps with sensible defaults
	applyDefaults(&cfg)

	return &cfg, nil
}

// decodeFile is a private helper to open and parse a single YAML file into the config struct.
func decodeFile(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	return dec.Decode(cfg)
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
}
