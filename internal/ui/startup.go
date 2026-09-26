package ui

import (
	"fmt"
	"io"

	"github.com/hareshkhan01/sticky-notes/internal/note"
)

// Startup renders each note in its own box for `stick startup`.
// Up to limit notes are shown, each with all of their topics.
func Startup(w io.Writer, o Options, notes []*note.Note) {
	t := o.Theme()
	if len(notes) == 0 {
		return
	}

	// Header line
	fmt.Fprintf(w, "%s %s\n\n",
		t.Title.Render("STICK"),
		t.Dim.Render("· "+itoa(len(notes))+" active"))

	for _, n := range notes {
		NotePage(w, o, n)
	}

	fmt.Fprintf(w, "%s\n", t.Dim.Render("stick list to see everything"))
}
