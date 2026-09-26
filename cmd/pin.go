package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newPinCmd builds `stick pin`.
func newPinCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "pin <id>",
		Short:             "Pin a note to the top of the list and startup panel",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPin(cmd, args[0], true)
		},
	}
}

// newUnpinCmd builds `stick unpin`.
func newUnpinCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "unpin <id>",
		Short:             "Unpin a note",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPin(cmd, args[0], false)
		},
	}
}

func runPin(cmd *cobra.Command, arg string, pinned bool) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()
	if err := c.open(); err != nil {
		return err
	}
	id, err := resolveID(arg)
	if err != nil {
		return err
	}
	if err := c.svc.SetPin(cmd.Context(), id, pinned); err != nil {
		return err
	}
	word := "pinned"
	if !pinned {
		word = "unpinned"
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Note "+word))
	return nil
}
