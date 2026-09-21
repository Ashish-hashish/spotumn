package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID    string `yaml:"client_id"`
	Port        int    `yaml:"port"`
	RedirectURI string `yaml:"redirect_uri"`
	ArtRenderer string `yaml:"art_renderer"` // "auto" (default: image if terminal supports it, else ansi), "image", "ansi"
}

func GetDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".spotumn"
	}
	dir := filepath.Join(home, ".config", "spotumn")
	_ = os.MkdirAll(dir, 0700)
	return dir
}

const (
	DefaultClientID = "d420a117a32841c2b3474932e49fb54b"
	DefaultPort     = 8989
)

func Load() (*Config, error) {
	cfg := &Config{
		ClientID:    DefaultClientID,
		Port:        DefaultPort,
		ArtRenderer: "auto",
	}

	configPath := filepath.Join(GetDir(), "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil && os.IsNotExist(err) {
		template := "# spotumn configuration (Optional - works out of the box)\nclient_id: \"[your_client_id]\"\nport: 8989\n# art_renderer: auto # auto, image, ansi\n"
		_ = os.WriteFile(configPath, []byte(template), 0600)
	} else if err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	// Environment variables take precedence over config file
	if envID := strings.TrimSpace(os.Getenv("SPOTUMN_CLIENT_ID")); envID != "" {
		cfg.ClientID = envID
	}
	if envURI := strings.TrimSpace(os.Getenv("SPOTUMN_REDIRECT_URI")); envURI != "" {
		cfg.RedirectURI = envURI
	}
	if envArt := strings.TrimSpace(os.Getenv("SPOTUMN_ART_RENDERER")); envArt != "" {
		cfg.ArtRenderer = strings.ToLower(envArt)
	}

	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	if cfg.ClientID == "" || cfg.ClientID == "[your_client_id]" {
		cfg.ClientID = DefaultClientID
	}
	cfg.RedirectURI = strings.TrimSpace(cfg.RedirectURI)
	if cfg.Port <= 0 {
		cfg.Port = DefaultPort
	}
	cfg.ArtRenderer = strings.ToLower(strings.TrimSpace(cfg.ArtRenderer))
	if cfg.ArtRenderer != "image" && cfg.ArtRenderer != "ansi" {
		cfg.ArtRenderer = "auto"
	}

	return cfg, nil
}
