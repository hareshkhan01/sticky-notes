package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

var startupFlags struct {
	limit int
}

// newStartupCmd builds `stick startup`, intended for shell startup hooks.
func newStartupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:    "startup",
		Short:  "Render the startup note panel (used by shell hooks)",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStartup(cmd)
		},
	}
	c.Flags().IntVarP(&startupFlags.limit, "limit", "n", 0, "override the configured note limit")
	return c
}

func runStartup(cmd *cobra.Command) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		// A broken config must never block a shell from opening.
		fmt.Fprintf(c.stderr, "stick: %v\n", err)
		return nil
	}

	cfg := c.cfg
	if !cfg.Startup.Enabled {
		return nil
	}
	limit := cfg.Startup.Limit
	if cmd.Flags().Changed("limit") {
		limit = startupFlags.limit
		if limit > 3 {
			limit = 3
		}
		c.cfg.Startup.Limit = limit
		if err := c.cfg.Save(c.cfgPath); err != nil {
			fmt.Fprintf(c.stderr, "stick: save config: %v\n", err)
		}
	}

	// Never render into a pipe or script; startup output is for humans
	// at an interactive prompt only.
	if !isInteractiveStdout() {
		return nil
	}

	if err := c.open(); err != nil {
		// Same rule for a broken or missing database: warn on stderr,
		// exit zero.
		fmt.Fprintf(c.stderr, "stick: %v\n", err)
		return nil
	}
	defer c.close()

	filter := note.NoteFilter{
		IncludeArchived: false,
		PinnedFirst:     true,
		Limit:           limit,
	}
	notes, err := c.svc.List(cmd.Context(), filter)
	if err != nil {
		fmt.Fprintf(c.stderr, "stick: %v\n", err)
		return nil
	}
	if cfg.Startup.PinnedOnly {
		notes = filterPinned(notes)
	}
	if len(notes) == 0 {
		if cfg.Startup.ShowWhenEmpty {
			fmt.Fprintln(c.stdout, c.ui.Theme().Dim.Render("stick: no notes yet — try 'stick add \"hello\"'"))
		}
		return nil
	}
	ui.Startup(c.stdout, c.ui, notes)
	return nil
}

// isInteractiveStdout reports whether stdout is a user terminal.
func isInteractiveStdout() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
