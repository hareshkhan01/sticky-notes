package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newConfigCmd builds `stick config` and its subcommands.
func newConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Show or change configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowConfig(cmd)
		},
	}
	c.AddCommand(newConfigStartupCmd())
	return c
}

func runShowConfig(cmd *cobra.Command) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()

	cfg := c.cfg
	if err := c.emitJSON(cfg); err != nil {
		return err
	}
	if flagJSON {
		return nil
	}
	t := c.ui.Theme()
	fmt.Fprintf(c.stdout, "startup.enabled        %s\n", onOff(cfg.Startup.Enabled))
	fmt.Fprintf(c.stdout, "startup.limit          %d\n", cfg.Startup.Limit)
	fmt.Fprintf(c.stdout, "startup.pinned_only    %s\n", onOff(cfg.Startup.PinnedOnly))
	fmt.Fprintf(c.stdout, "startup.show_when_empty %s\n", onOff(cfg.Startup.ShowWhenEmpty))
	fmt.Fprintf(c.stdout, "display.color          %s\n", cfg.Display.Color)
	fmt.Fprintf(c.stdout, "display.date_format    %s\n", cfg.Display.DateFormat)
	fmt.Fprintf(c.stdout, "\n%s\n", t.Dim.Render("config file: "+c.cfgPath))
	fmt.Fprintf(c.stdout, "%s\n", t.Dim.Render("database:    "+c.dbPath))
	return nil
}

// newConfigStartupCmd builds `stick config startup`.
func newConfigStartupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "startup",
		Short: "Enable or disable the startup panel",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowConfig(cmd)
		},
	}
	c.AddCommand(&cobra.Command{
		Use:               "enable",
		Short:             "Show notes when a terminal opens",
		Args:              cobra.NoArgs,
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setStartupEnabled(cmd, true)
		},
	}, &cobra.Command{
		Use:               "disable",
		Short:             "Do not show notes when a terminal opens",
		Args:              cobra.NoArgs,
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setStartupEnabled(cmd, false)
		},
	})
	return c
}

func setStartupEnabled(cmd *cobra.Command, enabled bool) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()

	c.cfg.Startup.Enabled = enabled
	if err := c.cfg.Save(c.cfgPath); err != nil {
		return err
	}
	state := "disabled"
	if enabled {
		state = "enabled"
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Startup panel "+state))
	return nil
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
