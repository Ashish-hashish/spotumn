package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID string `yaml:"client_id"`
	Port     int    `yaml:"port"`
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
		ClientID: DefaultClientID,
		Port:     DefaultPort,
	}

	if envID := strings.TrimSpace(os.Getenv("SPOTUMN_CLIENT_ID")); envID != "" {
		cfg.ClientID = envID
		return cfg, nil
	}

	configPath := filepath.Join(GetDir(), "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil && os.IsNotExist(err) {
		template := "# spotumn configuration (Optional - works out of the box)\nclient_id: \"[your_client_id]\"\nport: 8989\n"
		_ = os.WriteFile(configPath, []byte(template), 0600)
	} else if err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	if cfg.ClientID == "" || cfg.ClientID == "[your_client_id]" {
		cfg.ClientID = DefaultClientID
	}
	if cfg.Port <= 0 {
		cfg.Port = DefaultPort
	}

	return cfg, nil
}
