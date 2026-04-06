package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ColorblindMode bool `json:"colorblind_mode"`
}

func DefaultConfigPath() (string, error) {
	dir, err := DefaultSaveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() Config {
	path, err := DefaultConfigPath()
	if err != nil {
		return Config{}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}
	}
	var cfg Config
	json.Unmarshal(b, &cfg)
	return cfg
}

func SaveConfig(cfg Config) error {
	path, err := DefaultConfigPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
