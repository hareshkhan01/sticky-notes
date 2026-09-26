package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCommand executes the command tree with isolated config/data dirs and
// captures stdout and stderr.
func runCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	dir := os.Getenv("STICK_TEST_HOME")
	if dir == "" {
		// Create the isolated root once per test so successive runCommand
		// calls share the same config/data directories.
		dir = t.TempDir()
		t.Setenv("STICK_TEST_HOME", dir)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	flagDB = "" // reset globals mutated by prior tests

	var stdout, stderr bytes.Buffer
	root := New()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestAddListShowDeleteFlow(t *testing.T) {
	out, _, err := runCommand(t, "add", "--title", "Errands", "buy milk\npost letter")
	if err != nil {
		t.Fatalf("add: %v\n%s", err, out)
	}
	if !strings.Contains(out, "✓ Note added") || !strings.Contains(out, "Errands") {
		t.Fatalf("unexpected add output:\n%s", out)
	}

	out, _, err = runCommand(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "Errands") {
		t.Fatalf("list should contain the note:\n%s", out)
	}

	out, _, err = runCommand(t, "show", "1")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if !strings.Contains(out, "buy milk") || !strings.Contains(out, "post letter") {
		t.Fatalf("show should print topics:\n%s", out)
	}

	out, _, err = runCommand(t, "delete", "--yes", "1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !strings.Contains(out, "✓ Note deleted") {
		t.Fatalf("unexpected delete output:\n%s", out)
	}

	if _, _, err = runCommand(t, "show", "1"); err == nil {
		t.Fatal("expected error showing a deleted note")
	}
}

func TestAddRejectsSixTopics(t *testing.T) {
	_, stderr, err := runCommand(t, "add", "1\n2\n3\n4\n5\n6")
	if err == nil {
		t.Fatal("expected error for six topics")
	}
	if !strings.Contains(stderr+errMessage(t, err), "5 topics") {
		t.Fatalf("expected topic-limit error, got: %v %s", err, stderr)
	}
}

func TestAddRejectsLongTopic(t *testing.T) {
	long := strings.Repeat("x", 201)
	_, _, err := runCommand(t, "add", long)
	if err == nil || !strings.Contains(errMessage(t, err), "the limit is 200") {
		t.Fatalf("expected topic length error, got %v", err)
	}
}

func TestListJSONOutput(t *testing.T) {
	if _, _, err := runCommand(t, "add", "json probe"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runCommand(t, "--json", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\"topics\"") || strings.Contains(out, "┌") {
		t.Fatalf("JSON output should be plain and structured:\n%s", out)
	}
}

func TestQuickAddViaRoot(t *testing.T) {
	out, _, err := runCommand(t, "remember the milk")
	if err != nil {
		t.Fatalf("quick add: %v", err)
	}
	if !strings.Contains(out, "remember the milk") {
		t.Fatalf("quick add output should show the note:\n%s", out)
	}
	// Title is derived from the first topic.
	out, _, _ = runCommand(t, "list")
	if !strings.Contains(out, "remember the milk") {
		t.Fatalf("list should show derived title:\n%s", out)
	}
}

func TestSearchFindsNote(t *testing.T) {
	if _, _, err := runCommand(t, "add", "--title", "Docker", "learn compose"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runCommand(t, "search", "compose")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(out, "Docker") {
		t.Fatalf("search should find the note:\n%s", out)
	}
}

func TestStartupSilentWhenNoNotes(t *testing.T) {
	out, _, err := runCommand(t, "startup")
	if err != nil {
		t.Fatalf("startup must not fail: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("startup should print nothing without notes, got:\n%s", out)
	}
}

func TestVersionPrints(t *testing.T) {
	out, _, err := runCommand(t, "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "stick ") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

// errMessage is a tiny helper to stringify errors in assertions.
func errMessage(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		return ""
	}
	return err.Error()
}
