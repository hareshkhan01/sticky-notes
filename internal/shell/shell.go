// Package shell installs and removes a clearly delimited "managed block"
// in the user's shell startup file. Install is idempotent, always backs
// the file up first, and removal only ever deletes the managed block.
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Managed block delimiters. Everything between them belongs to Stick.
const (
	BeginMarker = "# >>> stick notes startup >>>"
	EndMarker   = "# <<< stick notes startup <<<"
)

// Kind identifies a supported shell.
type Kind string

// Supported shells.
const (
	Bash Kind = "bash"
	Zsh  Kind = "zsh"
	Fish Kind = "fish"
)

// ParseKind validates a shell name supplied on the command line.
func ParseKind(name string) (Kind, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	case "fish":
		return Fish, nil
	default:
		return "", fmt.Errorf("unsupported shell %q; want bash, zsh, or fish", name)
	}
}

// StartupFile returns the startup file path for a shell, allowing HOME to
// be overridden for tests.
func StartupFile(k Kind, home string) (string, error) {
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("shell: resolve home directory: %w", err)
		}
		home = h
	}
	switch k {
	case Bash:
		return filepath.Join(home, ".bashrc"), nil
	case Zsh:
		return filepath.Join(home, ".zshrc"), nil
	case Fish:
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("shell: unknown kind %q", k)
	}
}

// hook returns the managed block body for a shell.
func hook(k Kind) string {
	switch k {
	case Fish:
		return `# >>> stick notes startup >>>
if status is-interactive; and type -q stick
    stick startup
end
# <<< stick notes startup <<<`
	default: // bash and zsh share POSIX-compatible syntax here
		return `# >>> stick notes startup >>>
if [[ $- == *i* ]] && command -v stick >/dev/null 2>&1; then
    stick startup
fi
# <<< stick notes startup <<<`
	}
}

// ContainsBlock reports whether the file content already holds a complete
// managed block. An unbalanced file is reported as an error.
func ContainsBlock(content string) (bool, error) {
	b, e := strings.Contains(content, BeginMarker), strings.Contains(content, EndMarker)
	switch {
	case b && e:
		return true, nil
	case b || e:
		return false, fmt.Errorf("shell: startup file has an unbalanced stick block; fix it by hand")
	default:
		return false, nil
	}
}

// Install appends the managed block to path if it is missing. It writes a
// timestamped backup next to the config dir first and returns whether the
// file changed.
func Install(k Kind, path, backupDir string) (changed bool, err error) {
	if _, err := ParseKind(string(k)); err != nil {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("shell: create config directory: %w", err)
	}
	existing, err := os.ReadFile(path)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		existing = nil
	default:
		return false, fmt.Errorf("shell: read %s: %w", path, err)
	}
	content := string(existing)

	hasBlock, err := ContainsBlock(content)
	if err != nil {
		return false, err
	}
	if hasBlock {
		return false, nil
	}

	if len(existing) > 0 {
		if err := backup(backupDir, path, existing); err != nil {
			return false, err
		}
	}

	var out strings.Builder
	out.WriteString(content)
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		out.WriteString("\n")
	}
	out.WriteString("\n")
	out.WriteString(hook(k))
	out.WriteString("\n")

	if err := writeFile(path, []byte(out.String())); err != nil {
		return false, err
	}
	return true, nil
}

// Uninstall removes only the managed block. Files without a block are
// left untouched; unbalanced blocks abort with an error.
func Uninstall(path, backupDir string) (changed bool, err error) {
	content, err := os.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("shell: read %s: %w", path, err)
	}

	hasBlock, err := ContainsBlock(string(content))
	if err != nil {
		return false, err
	}
	if !hasBlock {
		return false, nil
	}

	if err := backup(backupDir, path, content); err != nil {
		return false, err
	}

	begin := strings.Index(string(content), BeginMarker)
	end := strings.Index(string(content), EndMarker)
	if begin < 0 || end < 0 || end < begin {
		return false, fmt.Errorf("shell: cannot locate block boundaries in %s", path)
	}
	after := end + len(EndMarker)
	cleaned := string(content)[:begin] + string(content)[after:]
	cleaned = strings.TrimRight(cleaned, "\n") + "\n"

	if err := writeFile(path, []byte(cleaned)); err != nil {
		return false, err
	}
	return true, nil
}

// writeFile creates-or-truncates path with mode 0o644 (0o755 not needed
// for a sourced config) without following symlinks for the final write.
func writeFile(path string, data []byte) error {
	tmp := path + ".stick-tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("shell: write %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("shell: replace %s: %w", path, err)
	}
	return nil
}

// backup copies original into backupDir with a timestamped name.
func backup(backupDir, path string, original []byte) error {
	if backupDir == "" {
		return nil
	}
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return fmt.Errorf("shell: create backup directory: %w", err)
	}
	name := fmt.Sprintf("%s.stick-backup-%s", filepath.Base(path), time.Now().Format("20060102-150405"))
	dst := filepath.Join(backupDir, name)
	if err := os.WriteFile(dst, original, 0o600); err != nil {
		return fmt.Errorf("shell: write backup: %w", err)
	}
	return nil
}
