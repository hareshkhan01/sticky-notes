package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteFlags struct {
	yes bool
}

// newDeleteCmd builds `stick delete`.
func newDeleteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:               "delete <id>",
		Aliases:           []string{"rm"},
		Short:             "Delete a note permanently",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args[0])
		},
	}
	c.Flags().BoolVarP(&deleteFlags.yes, "yes", "y", false, "skip confirmation")
	return c
}

func runDelete(cmd *cobra.Command, arg string) error {
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
	n, err := c.svc.Get(cmd.Context(), id)
	if err != nil {
		return err
	}

	ok, err := c.confirm(fmt.Sprintf("Delete note %d (%s)?", id, displayTitle(n)), deleteFlags.yes)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(c.stdout, "Cancelled.")
		return nil
	}
	if err := c.svc.Delete(cmd.Context(), id); err != nil {
		return err
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Note deleted"))
	return nil
}

// newArchiveCmd builds `stick archive` and `stick unarchive`.
func newArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "archive <id>",
		Short:             "Archive a note (hidden from the default list)",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchive(cmd, args[0], true)
		},
	}
}

// newUnarchiveCmd builds `stick unarchive`.
func newUnarchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "unarchive <id>",
		Short:             "Restore an archived note",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchive(cmd, args[0], false)
		},
	}
}

func runArchive(cmd *cobra.Command, arg string, archived bool) error {
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
	if err := c.svc.SetArchive(cmd.Context(), id, archived); err != nil {
		return err
	}
	word := "archived"
	if !archived {
		word = "restored"
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Note "+word))
	return nil
}
