package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/config"
	"github.com/hareshkhan01/sticky-notes/internal/shell"
)

var initFlags struct {
	yes    bool
	backup bool
}

// newInitCmd builds `stick init <shell>`.
func newInitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "init <bash|zsh|fish>",
		Short: "Install the terminal-startup hook for a shell",
		Long: `Install the shell hook that shows your notes when a terminal opens.

Stick appends a clearly marked block to your shell startup file after
showing you what will change and asking for confirmation. A timestamped
backup is written to the stick config directory first. Running init twice
never duplicates the block.

By default, only the latest note is shown at startup. Use
'stick startup --limit N' to change this later.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeShells,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args[0])
		},
	}
	c.Flags().BoolVarP(&initFlags.yes, "yes", "y", false, "skip the confirmation prompt")
	return c
}

func runInit(cmd *cobra.Command, name string) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()

	kind, err := shell.ParseKind(name)
	if err != nil {
		return err
	}
	path, err := shell.StartupFile(kind, "")
	if err != nil {
		return err
	}
	backupDir := filepath.Join(config.DataDir(), "backups")

	fmt.Fprintf(c.stdout, "Shell:   %s\n", kind)
	fmt.Fprintf(c.stdout, "File:    %s\n", path)
	fmt.Fprintf(c.stdout, "Block:   %s\n", shell.BeginMarker)
	fmt.Fprintf(c.stdout, "Startup: only the latest note\n")
	fmt.Fprintf(c.stdout, "Effect:  run 'stick startup' in interactive shells\n\n")

	ok, err := c.confirm(fmt.Sprintf("Append the stick block to %s?", path), initFlags.yes)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(c.stdout, "Cancelled; your shell configuration was not modified.")
		return nil
	}

	changed, err := shell.Install(kind, path, backupDir)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Already installed; nothing changed."))
		return nil
	}
	c.cfg.Startup.Limit = 1
	if err := c.cfg.Save(c.cfgPath); err != nil {
		return err
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Hook installed."))
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Startup limit set to 1 (latest note only)"))
	fmt.Fprintf(c.stdout, "Open a new terminal or run 'source %s' to see your notes.\n", path)
	return nil
}

var uninstallFlags struct{ yes bool }

// newUninstallShellCmd builds `stick uninstall-shell <shell>`.
func newUninstallShellCmd() *cobra.Command {
	c := &cobra.Command{
		Use:               "uninstall-shell <bash|zsh|fish>",
		Short:             "Remove the stick block from a shell startup file",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeShells,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstallShell(cmd, args[0])
		},
	}
	c.Flags().BoolVarP(&uninstallFlags.yes, "yes", "y", false, "skip the confirmation prompt")
	return c
}

func runUninstallShell(cmd *cobra.Command, name string) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()

	kind, err := shell.ParseKind(name)
	if err != nil {
		return err
	}
	path, err := shell.StartupFile(kind, "")
	if err != nil {
		return err
	}
	backupDir := filepath.Join(config.DataDir(), "backups")

	ok, err := c.confirm(fmt.Sprintf("Remove the stick block from %s?", path), uninstallFlags.yes)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(c.stdout, "Cancelled.")
		return nil
	}

	changed, err := shell.Uninstall(path, backupDir)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Fprintln(c.stdout, "No stick block found; nothing changed.")
		return nil
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Hook removed. Your notes were not touched."))
	return nil
}

// completeShells offers shell names for completion.
func completeShells(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{"bash", "zsh", "fish"}, cobra.ShellCompDirectiveNoFileComp
}
