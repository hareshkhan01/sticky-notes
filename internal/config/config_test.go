package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsWhenMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	want := Default()
	if cfg != want {
		t.Fatalf("expected defaults %+v, got %+v", want, cfg)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg := Default()
	cfg.Startup.Limit = 3
	cfg.Startup.PinnedOnly = true
	cfg.Display.Color = "never"
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != cfg {
		t.Fatalf("round trip mismatch: %+v vs %+v", got, cfg)
	}
}

func TestLoadRepairsInvalidValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	broken := "startup:\n  limit: -5\n  enabled: true\ndisplay:\n  color: sometimes\n"
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Startup.Limit != Default().Startup.Limit {
		t.Fatalf("expected limit repaired to %d, got %d", Default().Startup.Limit, cfg.Startup.Limit)
	}
	if cfg.Display.Color != "auto" {
		t.Fatalf("expected color repaired to auto, got %q", cfg.Display.Color)
	}
}

func TestLoadRejectsInvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("startup: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected parse error for invalid YAML")
	}
}

func TestSaveIsAtomic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := Default()
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temp file should not survive a successful save")
	}
}
