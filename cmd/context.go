// Package cmd defines the Cobra command tree for Stick. Commands parse
// flags, call the note service, and delegate all rendering to ui.
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/hareshkhan01/sticky-notes/internal/config"
	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/storage"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

// cli carries everything a command needs during one invocation.
type cli struct {
	stdout  io.Writer
	stderr  io.Writer
	stdin   io.Reader
	ui      ui.Options
	cfg     config.Config
	cfgPath string
	dbPath  string
	store   *storage.Store
	svc     *note.Service
}

// Globals bound to root-command flags.
var (
	flagDB      string
	flagNoColor bool
	flagJSON    bool
)

// newCLI resolves configuration and opens the database lazily.
func newCLI(stdout, stderr io.Writer, stdin io.Reader) *cli {
	return &cli{stdout: stdout, stderr: stderr, stdin: stdin}
}

// setup loads configuration and applies global output flags. It does not
// touch the database; commands call open() when they need it.
func (c *cli) setup() error {
	c.cfgPath = config.Path()
	cfg, err := config.Load(c.cfgPath)
	if err != nil {
		return err
	}
	c.cfg = cfg
	c.dbPath = flagDB
	if c.dbPath == "" {
		c.dbPath = config.DBPath()
	}
	c.ui = ui.Options{
		NoColor:    flagNoColor,
		JSON:       flagJSON,
		ColorMode:  cfg.Display.Color,
		DateFormat: cfg.Display.DateFormat,
	}
	return nil
}

// open initialises storage and the note service.
func (c *cli) open() error {
	if c.store != nil {
		return nil
	}
	st, err := storage.Open(c.dbPath)
	if err != nil {
		return err
	}
	c.store = st
	c.svc = note.NewService(st)
	return nil
}

// close releases storage; safe to call more than once.
func (c *cli) close() {
	if c.svc != nil {
		c.svc.Close()
		c.svc = nil
		c.store = nil
	}
}

// emitJSON writes v as indented JSON when --json is set.
func (c *cli) emitJSON(v any) error {
	if !flagJSON {
		return nil
	}
	enc := json.NewEncoder(c.stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// confirm asks the user before a destructive action. When stdin is not a
// terminal and --yes was not passed, it refuses rather than guessing.
func (c *cli) confirm(prompt string, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		return false, errors.New("refusing to proceed without confirmation; pass --yes for scripted use")
	}
	return ui.Confirm(c.stdin, c.stdout, prompt), nil
}

// resolveID parses a note ID argument.
func resolveID(arg string) (int64, error) {
	var id int64
	if _, err := fmt.Sscanf(arg, "%d", &id); err != nil || id <= 0 {
		return 0, fmt.Errorf("%q is not a valid note ID", arg)
	}
	return id, nil
}
