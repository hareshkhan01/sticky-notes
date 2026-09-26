package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

var addFlags struct {
	title string
	pin   bool
}

// newAddCmd builds `stick add`.
func newAddCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "add [text]",
		Short: "Create a new sticky note",
		Long: `Create a new sticky note.

Each note is a page holding up to 5 topics, one topic per line, with a
200-character limit per topic. When --title is omitted, the first topic
becomes the title.

Examples:
  stick add "buy milk"
  stick add --title "Errands" "buy milk
  pick up dry cleaning"
  printf 'topic one\ntopic two\n' | stick add --title "Today" -`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd, args)
		},
	}
	c.Flags().StringVarP(&addFlags.title, "title", "t", "", "note title (max 60 characters)")
	c.Flags().BoolVarP(&addFlags.pin, "pin", "p", false, "pin the new note")
	return c
}

func runAdd(cmd *cobra.Command, args []string) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()
	if err := c.open(); err != nil {
		return err
	}

	var raw string
	switch {
	case len(args) == 1 && args[0] == "-":
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		raw = string(data)
	case len(args) == 1:
		raw = args[0]
	default:
		return errors.New("provide note text or use --title; try 'stick add --help'")
	}

	n, err := c.svc.Create(cmd.Context(), addFlags.title, raw)
	if err != nil {
		switch {
		case errors.Is(err, note.ErrTooManyTopics):
			return fmt.Errorf("%w (split large captures into multiple notes)", err)
		case errors.Is(err, note.ErrEmpty):
			return errors.New("note text is empty; nothing was saved")
		default:
			return err
		}
	}
	if addFlags.pin {
		if err := c.svc.SetPin(cmd.Context(), n.ID, true); err != nil {
			return err
		}
		n.IsPinned = true
	}

	if err := c.emitJSON(noteToJSON(n)); err != nil {
		return err
	}
	if flagJSON {
		return nil
	}

	t := c.ui.Theme()
	fmt.Fprintln(c.stdout)
	fmt.Fprintln(c.stdout, t.Ok.Render("✓ Note added"))
	ui.NotePage(c.stdout, c.ui, n)
	return nil
}
