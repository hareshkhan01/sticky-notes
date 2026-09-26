// Package config loads and persists user preferences, deriving
// OS-appropriate default paths for the database and configuration file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Startup holds preferences for the shell-startup note panel.
type Startup struct {
	Enabled       bool `yaml:"enabled"`
	Limit         int  `yaml:"limit"`
	PinnedOnly    bool `yaml:"pinned_only"`
	ShowWhenEmpty bool `yaml:"show_when_empty"`
}

// Display holds rendering preferences.
type Display struct {
	Color      string `yaml:"color"`       // "auto", "always", "never"
	DateFormat string `yaml:"date_format"` // Go reference layout
}

// Config is the on-disk configuration document.
type Config struct {
	Startup Startup `yaml:"startup"`
	Display Display `yaml:"display"`
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{
		Startup: Startup{
			Enabled:       true,
			Limit:         3,
			PinnedOnly:    false,
			ShowWhenEmpty: false,
		},
		Display: Display{
			Color:      "auto",
			DateFormat: "2006-01-02 15:04",
		},
	}
}

// Dir returns the configuration directory, honouring XDG on Unix.
func Dir() string {
	if runtime.GOOS == "windows" {
		if d, err := os.UserConfigDir(); err == nil && d != "" {
			return filepath.Join(d, "stick")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "stick")
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "stick")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "stick"
	}
	return filepath.Join(home, ".config", "stick")
}

// Path returns the configuration file path.
func Path() string { return filepath.Join(Dir(), "config.yaml") }

// DataDir returns the directory holding the database.
func DataDir() string {
	if runtime.GOOS == "windows" {
		if d, err := os.UserCacheDir(); err == nil && d != "" {
			return filepath.Join(d, "stick")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "stick")
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "stick")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "stick"
	}
	return filepath.Join(home, ".local", "share", "stick")
}

// DBPath returns the default database file path.
func DBPath() string { return filepath.Join(DataDir(), "notes.db") }

// Load reads the config file, falling back to defaults when missing.
// Invalid YAML is an error; invalid *values* are repaired to defaults.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse %s: %w", path, err)
	}
	cfg.normalize()
	return cfg, nil
}

// Save writes the config atomically (temp file plus rename).
func (c Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("config: create directory: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("config: encode: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("config: write: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("config: rename: %w", err)
	}
	return nil
}

// normalize repairs out-of-range values instead of failing at startup.
func (c *Config) normalize() {
	if c.Startup.Limit <= 0 || c.Startup.Limit > 3 {
		c.Startup.Limit = Default().Startup.Limit
	}
	switch strings.ToLower(c.Display.Color) {
	case "always", "never":
	case "auto", "":
		c.Display.Color = "auto"
	default:
		c.Display.Color = "auto"
	}
	if c.Display.DateFormat == "" {
		c.Display.DateFormat = Default().Display.DateFormat
	}
}
