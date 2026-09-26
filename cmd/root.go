package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

// version information, overridable at link time with -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// New builds the complete Stick command tree.
func New() *cobra.Command {
	root := &cobra.Command{
		Use:   "stick",
		Short: "Sticky notes for your terminal",
		Long: `Stick — sticky notes for your terminal.

Each note is a page holding up to 5 topics with a 200-character limit
per topic. Notes live in a local SQLite database; nothing leaves your
machine.

Quick capture:
  stick "remember the milk"
  stick add --title "Errands" "buy milk
  pick up dry cleaning"`,
		// Errors are printed once by Execute (below); usage stays quiet so
		// the message is never buried.
		SilenceUsage:  true,
		SilenceErrors: true,
		// A bare positional argument is treated as a quick-add note.
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("accepts at most 1 arg for quick capture")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return quickAdd(cmd, args[0])
			}
			return cmd.Help()
		},
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&flagDB, "db", "", "", "use an alternate database path")
	pf.BoolVarP(&flagNoColor, "no-color", "", false, "disable colored output")
	pf.BoolVarP(&flagJSON, "json", "", false, "print machine-readable JSON")

	root.AddCommand(
		newAddCmd(),
		newListCmd(),
		newShowCmd(),
		newEditCmd(),
		newDeleteCmd(),
		newSearchCmd(),
		newPinCmd(),
		newUnpinCmd(),
		newArchiveCmd(),
		newUnarchiveCmd(),
		newStartupCmd(),
		newConfigCmd(),
		newInitCmd(),
		newUninstallShellCmd(),
		newDoctorCmd(),
		newVersionCmd(),
	)
	return root
}

// Execute runs the command tree and returns the process exit code. It
// prints every error exactly once, to stderr, with an actionable hint
// where one exists.
func Execute() int {
	if err := New().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "stick: %v\n", err)
		var nf note.NotFoundError
		if errors.As(err, &nf) {
			fmt.Fprintln(os.Stderr, "stick: run 'stick list' to see note IDs")
		}
		return 1
	}
	return 0
}

// quickAdd handles `stick "some text"` by delegating to the add flow.
func quickAdd(cmd *cobra.Command, text string) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()
	if err := c.open(); err != nil {
		return err
	}
	n, err := c.svc.Create(cmd.Context(), "", text)
	if err != nil {
		if errors.Is(err, note.ErrTooManyTopics) {
			return fmt.Errorf("%w (split large captures into multiple notes)", err)
		}
		return err
	}
	if err := c.emitJSON(noteToJSON(n)); err != nil {
		return err
	}
	if flagJSON {
		return nil
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Note added"))
	ui.NotePage(c.stdout, c.ui, n)
	return nil
}

// newVersionCmd builds `stick version`.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "stick %s (commit %s, built %s)\n", version, commit, date)
			return nil
		},
	}
}

// noCompletions disables file completion for commands that take IDs.
func noCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
