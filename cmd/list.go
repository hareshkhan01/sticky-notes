package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

var listFlags struct {
	all   bool
	limit int
	pin   bool
}

// newListCmd builds `stick list`.
func newListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List sticky notes",
		Long: `List sticky notes in a compact table.

Pinned notes appear first. Archived notes are hidden unless --all is given.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd)
		},
	}
	c.Flags().BoolVar(&listFlags.all, "all", false, "include archived notes")
	c.Flags().IntVarP(&listFlags.limit, "limit", "n", 0, "show at most n notes (0 = all)")
	c.Flags().BoolVarP(&listFlags.pin, "pin", "p", false, "show only pinned notes")
	return c
}

func runList(cmd *cobra.Command) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()
	if err := c.open(); err != nil {
		return err
	}

	filter := note.NoteFilter{
		IncludeArchived: listFlags.all,
		PinnedFirst:     true,
		Limit:           listFlags.limit,
	}
	if listFlags.pin {
		filter.IncludeArchived = false
	}

	notes, err := c.svc.List(cmd.Context(), filter)
	if err != nil {
		return err
	}
	if listFlags.pin {
		notes = filterPinned(notes)
	}

	total, err := c.store.CountActive(cmd.Context())
	if err != nil {
		return err
	}

	if err := c.emitJSON(notesToJSON(notes)); err != nil {
		return err
	}
	if flagJSON {
		return nil
	}

	rows := make([]ui.NoteRow, len(notes))
	for i, n := range notes {
		rows[i] = ui.NoteRow{
			ID:      n.ID,
			Title:   displayTitle(n),
			Pinned:  n.IsPinned,
			Topics:  len(n.Topics),
			Updated: ui.FmtTime(c.ui, n.UpdatedAt),
		}
	}
	ui.List(c.stdout, c.ui, rows, total)

	if len(rows) > 0 {
		fmt.Fprintf(c.stdout, "\nUse 'stick show <id>' to view a note, or 'stick add' to create one.\n")
	}
	return nil
}

// filterPinned keeps only pinned notes, preserving order.
func filterPinned(notes []*note.Note) []*note.Note {
	out := notes[:0]
	for _, n := range notes {
		if n.IsPinned {
			out = append(out, n)
		}
	}
	return out
}

// displayTitle falls back to a placeholder for untitled notes.
func displayTitle(n *note.Note) string {
	if n.Title != "" {
		return n.Title
	}
	return "(untitled)"
}
