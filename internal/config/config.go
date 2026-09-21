package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SpotifyClientID is the fixed built-in Spotify Connect client ID
const SpotifyClientID = "d420a117a32841c2b3474932e49fb54b"

type Config struct {
	Port        int    `yaml:"port"`
	RedirectURI string `yaml:"redirect_uri"`
	ArtRenderer string `yaml:"art_renderer"` // "auto", "image", "ansi"
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
	DefaultPort = 8989
)

func Load() (*Config, error) {
	cfg := &Config{
		Port:        DefaultPort,
		ArtRenderer: "auto",
	}

	configPath := filepath.Join(GetDir(), "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil && os.IsNotExist(err) {
		template := "# spotumn configuration\nport: 8989\n# art_renderer: auto # auto, ansi\n"
		_ = os.WriteFile(configPath, []byte(template), 0600)
	} else if err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	// Environment variables take precedence over config file
	if envURI := strings.TrimSpace(os.Getenv("SPOTUMN_REDIRECT_URI")); envURI != "" {
		cfg.RedirectURI = envURI
	}
	if envArt := strings.TrimSpace(os.Getenv("SPOTUMN_ART_RENDERER")); envArt != "" {
		cfg.ArtRenderer = strings.ToLower(envArt)
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
