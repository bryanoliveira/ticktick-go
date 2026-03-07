package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Timezone       string `json:"timezone"`
	DefaultProject string `json:"default_project"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
}

var defaultConfig = Config{
	Timezone:       "Europe/London",
	DefaultProject: "inbox",
	ClientID:       "24qE700R7e12YnSNWj",
	ClientSecret:   "4kF89Zm77tWhMvhNq0TiL4PTxavRTdCJ",
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tt", "config.json")
}

func Load() *Config {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return &defaultConfig
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &defaultConfig
	}

	if cfg.ClientID == "" {
		cfg.ClientID = defaultConfig.ClientID
	}
	if cfg.ClientSecret == "" {
		cfg.ClientSecret = defaultConfig.ClientSecret
	}
	if cfg.Timezone == "" {
		cfg.Timezone = defaultConfig.Timezone
	}

	return &cfg
}

func EnsureConfigDir() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "tt")
	return os.MkdirAll(dir, 0755)
}
