package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseKind(t *testing.T) {
	for _, ok := range []string{"bash", "BASH", "zsh", "fish"} {
		if _, err := ParseKind(ok); err != nil {
			t.Errorf("ParseKind(%q) = %v, want nil", ok, err)
		}
	}
	if _, err := ParseKind("tcsh"); err == nil {
		t.Error("expected error for unsupported shell")
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".zshrc")
	backupDir := filepath.Join(t.TempDir(), "backups")

	// Pre-existing user content must be preserved.
	userContent := "# my aliases\nalias ll='ls -la'\n"
	if err := os.WriteFile(path, []byte(userContent), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := Install(Zsh, path, backupDir)
	if err != nil || !changed {
		t.Fatalf("first install: changed=%v err=%v", changed, err)
	}

	changed, err = Install(Zsh, path, backupDir)
	if err != nil || changed {
		t.Fatalf("second install should be a no-op: changed=%v err=%v", changed, err)
	}

	data, _ := os.ReadFile(path)
	s := string(data)
	if !strings.HasPrefix(s, userContent) {
		t.Fatalf("user content was not preserved:\n%s", s)
	}
	if strings.Count(s, BeginMarker) != 1 || strings.Count(s, EndMarker) != 1 {
		t.Fatalf("expected exactly one managed block:\n%s", s)
	}
}

func TestInstallCreatesFishConfig(t *testing.T) {
	home := t.TempDir()
	backupDir := t.TempDir()
	path := filepath.Join(home, ".config", "fish", "config.fish")

	changed, err := Install(Fish, path, backupDir)
	if err != nil || !changed {
		t.Fatalf("install: changed=%v err=%v", changed, err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "status is-interactive") {
		t.Fatalf("fish hook missing interactive guard:\n%s", data)
	}
	if strings.Contains(string(data), "[[ $- == *i* ]]") {
		t.Fatal("fish hook must not contain bash syntax")
	}
}

func TestUninstallRemovesOnlyManagedBlock(t *testing.T) {
	home := t.TempDir()
	backupDir := t.TempDir()
	path := filepath.Join(home, ".bashrc")

	userContent := "export EDITOR=vim\n"
	if err := os.WriteFile(path, []byte(userContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(Bash, path, backupDir); err != nil {
		t.Fatal(err)
	}

	changed, err := Uninstall(path, backupDir)
	if err != nil || !changed {
		t.Fatalf("uninstall: changed=%v err=%v", changed, err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if strings.Contains(s, BeginMarker) || strings.Contains(s, "stick startup") {
		t.Fatalf("managed block not fully removed:\n%s", s)
	}
	if !strings.Contains(s, userContent) {
		t.Fatalf("user content lost:\n%s", s)
	}

	// Uninstalling again is a no-op.
	changed, err = Uninstall(path, backupDir)
	if err != nil || changed {
		t.Fatalf("second uninstall should be a no-op: changed=%v err=%v", changed, err)
	}
}

func TestUninstallRefusesUnbalancedBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(path, []byte(BeginMarker+"\nstick startup\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(path, t.TempDir()); err == nil {
		t.Fatal("expected error for unbalanced block")
	}
}

func TestInstallWritesBackup(t *testing.T) {
	home := t.TempDir()
	backupDir := t.TempDir()
	path := filepath.Join(home, ".zshrc")
	if err := os.WriteFile(path, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(Zsh, path, backupDir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one backup file, got %d (err=%v)", len(entries), err)
	}
	data, _ := os.ReadFile(filepath.Join(backupDir, entries[0].Name()))
	if string(data) != "original\n" {
		t.Fatalf("backup content mismatch: %q", data)
	}
}
