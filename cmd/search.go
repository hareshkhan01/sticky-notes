package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

var searchFlags struct {
	all   bool
	limit int
}

// newSearchCmd builds `stick search`.
func newSearchCmd() *cobra.Command {
	c := &cobra.Command{
		Use:               "search <query>",
		Short:             "Search note titles and topics",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, args[0])
		},
	}
	c.Flags().BoolVar(&searchFlags.all, "all", false, "include archived notes")
	c.Flags().IntVarP(&searchFlags.limit, "limit", "n", 0, "show at most n results (0 = all)")
	return c
}

func runSearch(cmd *cobra.Command, query string) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()
	if err := c.open(); err != nil {
		return err
	}
	notes, err := c.svc.Search(cmd.Context(), query, note.NoteFilter{
		IncludeArchived: searchFlags.all,
		PinnedFirst:     true,
		Limit:           searchFlags.limit,
	})
	if err != nil {
		return err
	}
	if err := c.emitJSON(notesToJSON(notes)); err != nil {
		return err
	}
	if flagJSON {
		return nil
	}
	if len(notes) == 0 {
		return fmt.Errorf("no notes match %q", query)
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
	ui.List(c.stdout, c.ui, rows, len(notes))
	return nil
}
