package cmd

import (
	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

// newShowCmd builds `stick show`.
func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "show <id>",
		Short:             "Show a note with all of its topics",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
			if err := c.setup(); err != nil {
				return err
			}
			defer c.close()
			if err := c.open(); err != nil {
				return err
			}
			id, err := resolveID(args[0])
			if err != nil {
				return err
			}
			n, err := c.svc.Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			if err := c.emitJSON(noteToJSON(n)); err != nil {
				return err
			}
			if flagJSON {
				return nil
			}
			ui.NotePage(c.stdout, c.ui, n)
			return nil
		},
	}
}
