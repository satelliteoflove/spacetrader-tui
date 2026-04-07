package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AnimSpeed int

const (
	AnimSlow   AnimSpeed = 1
	AnimMedium AnimSpeed = 2
	AnimFast   AnimSpeed = 3
	AnimOff    AnimSpeed = 4
)

func (a AnimSpeed) String() string {
	switch a {
	case AnimSlow:
		return "Slow"
	case AnimMedium:
		return "Medium"
	case AnimFast:
		return "Fast"
	case AnimOff:
		return "Off"
	default:
		return "Medium"
	}
}

func (a AnimSpeed) Next() AnimSpeed {
	switch a {
	case AnimSlow:
		return AnimMedium
	case AnimMedium:
		return AnimFast
	case AnimFast:
		return AnimOff
	case AnimOff:
		return AnimSlow
	default:
		return AnimMedium
	}
}

func (a AnimSpeed) Prev() AnimSpeed {
	switch a {
	case AnimSlow:
		return AnimOff
	case AnimMedium:
		return AnimSlow
	case AnimFast:
		return AnimMedium
	case AnimOff:
		return AnimFast
	default:
		return AnimMedium
	}
}

type Config struct {
	ColorblindMode    bool      `json:"colorblind_mode"`
	TransitionSpeed   AnimSpeed `json:"transition_speed"`
	WarpSpeed         AnimSpeed `json:"warp_speed"`
	EncounterEntrance AnimSpeed `json:"encounter_entrance"`
	TypewriterSpeed   AnimSpeed `json:"typewriter_speed"`
	PulseSpeed        AnimSpeed `json:"pulse_speed"`
}

func (c *Config) applyDefaults() {
	if c.TransitionSpeed == 0 {
		c.TransitionSpeed = AnimMedium
	}
	if c.WarpSpeed == 0 {
		c.WarpSpeed = AnimMedium
	}
	if c.EncounterEntrance == 0 {
		c.EncounterEntrance = AnimMedium
	}
	if c.TypewriterSpeed == 0 {
		c.TypewriterSpeed = AnimMedium
	}
	if c.PulseSpeed == 0 {
		c.PulseSpeed = AnimMedium
	}
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
	cfg.applyDefaults()
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
