package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/rivo/uniseg"
)

// NoteRow is one row of the list table.
type NoteRow struct {
	ID      int64
	Title   string
	Pinned  bool
	Topics  int
	Updated string // pre-formatted
}

// List renders the `stick list` table. All cells are padded to computed
// column widths so the result lines up in any terminal.
func List(w io.Writer, o Options, rows []NoteRow, total int) {
	t := o.Theme()

	idW, stW, tpW := len("ID"), len("ST"), len("TOPICS")
	titles := make([]string, len(rows))
	for i, r := range rows {
		titles[i] = r.Title
	}
	tiW := min(maxWidth(titles, len("TITLE")), max(12, o.width()-idW-stW-tpW-12))

	// Header
	header := pad("ID", idW) + "  " + pad("ST", stW) + "  " + pad("TITLE", tiW) + "  " +
		pad("TOPICS", tpW) + "  " + "UPDATED"
	sep := strings.Repeat("─", uniseg.StringWidth(header))
	fmt.Fprintln(w, t.Title.Render(header))
	fmt.Fprintln(w, t.Dim.Render(sep))

	if len(rows) == 0 {
		fmt.Fprintln(w, t.Dim.Render("no notes yet — try 'stick add \"hello world\"'"))
		return
	}

	for i, r := range rows {
		st := "  "
		if r.Pinned {
			st = t.PinMark
		}
		title := truncate(r.Title, tiW)
		line := pad(itoa(int(r.ID)), idW) + "  " +
			pad(st, stW) + "  " +
			pad(title, tiW) + "  " +
			pad(itoa(r.Topics), tpW) + "  " +
			r.Updated
		if i%2 == 1 {
			fmt.Fprintln(w, t.Dim.Render(line))
		} else {
			fmt.Fprintln(w, line)
		}
	}

	fmt.Fprintln(w, t.Dim.Render(sep))
	fmt.Fprintf(w, t.Dim.Render("%d of %d notes")+"\n", len(rows), total)
}

// maxWidth returns the widest display width in the slice.
func maxWidth(ss []string, floor int) int {
	m := floor
	for _, s := range ss {
		if w := uniseg.StringWidth(s); w > m {
			m = w
		}
	}
	return m
}
