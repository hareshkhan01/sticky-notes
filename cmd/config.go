package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/config"
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
	}, &cobra.Command{
		Use:               "pinned_only <on|off>",
		Short:             "Show only pinned notes at startup",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeOnOff,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setStartupBool(cmd, "pinned_only", args[0], func(c *config.Config, v bool) { c.Startup.PinnedOnly = v })
		},
	}, &cobra.Command{
		Use:               "show_when_empty <on|off>",
		Short:             "Show a hint when there are no notes",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeOnOff,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setStartupBool(cmd, "show_when_empty", args[0], func(c *config.Config, v bool) { c.Startup.ShowWhenEmpty = v })
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

func setStartupBool(cmd *cobra.Command, name, value string, set func(*config.Config, bool)) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()

	var v bool
	switch value {
	case "on", "true", "1":
		v = true
	case "off", "false", "0":
		v = false
	default:
		return fmt.Errorf("expected on or off, got %q", value)
	}
	set(&c.cfg, v)
	if err := c.cfg.Save(c.cfgPath); err != nil {
		return err
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ "+name+" "+onOff(v)))
	return nil
}

func completeOnOff(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"on\tenable", "off\tdisable"}, cobra.ShellCompDirectiveNoFileComp
}
