package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/shell"
)

// newDoctorCmd builds `stick doctor`.
func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check database, configuration, and shell integration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd)
		},
	}
}

func runDoctor(cmd *cobra.Command) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()

	ok := c.ui.Theme().Ok.Render("ok")
	bad := c.ui.Theme().Warn.Render("problem")

	// Database.
	if err := c.open(); err != nil {
		fmt.Fprintf(c.stdout, "%s  database %s: %v\n", bad, c.dbPath, err)
	} else {
		err := c.store.Ping(context.Background())
		if err != nil {
			fmt.Fprintf(c.stdout, "%s  database %s: %v\n", bad, c.dbPath, err)
		} else {
			fmt.Fprintf(c.stdout, "%s  database %s\n", ok, c.dbPath)
		}
	}

	// Configuration.
	if _, err := os.Stat(c.cfgPath); os.IsNotExist(err) {
		fmt.Fprintf(c.stdout, "%s  config %s (missing; defaults in use)\n", ok, c.cfgPath)
	} else {
		fmt.Fprintf(c.stdout, "%s  config %s\n", ok, c.cfgPath)
	}

	// Shell integration, all supported shells.
	for _, k := range []shell.Kind{shell.Bash, shell.Zsh, shell.Fish} {
		path, err := shell.StartupFile(k, "")
		if err != nil {
			fmt.Fprintf(c.stdout, "%s  %s: %v\n", bad, k, err)
			continue
		}
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			fmt.Fprintf(c.stdout, "%s  %s: no startup file at %s\n", ok, k, path)
			continue
		}
		if err != nil {
			fmt.Fprintf(c.stdout, "%s  %s: %v\n", bad, k, err)
			continue
		}
		has, err := shell.ContainsBlock(string(data))
		switch {
		case err != nil:
			fmt.Fprintf(c.stdout, "%s  %s: %v\n", bad, k, err)
		case has:
			fmt.Fprintf(c.stdout, "%s  %s: hook installed in %s\n", ok, k, filepath.Base(path))
		default:
			fmt.Fprintf(c.stdout, "%s  %s: no hook (run 'stick init %s' to add)\n", ok, k, k)
		}
	}

	// Startup preference.
	if !c.cfg.Startup.Enabled {
		fmt.Fprintf(c.stdout, "%s  startup panel disabled (run 'stick config startup enable')\n", bad)
	} else {
		fmt.Fprintf(c.stdout, "%s  startup panel enabled, limit %d\n", ok, c.cfg.Startup.Limit)
	}
	fmt.Fprintln(c.stdout)
	return nil
}
