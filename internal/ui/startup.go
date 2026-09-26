package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/hareshkhan01/sticky-notes/internal/note"
)

// Startup renders the compact panel used by `stick startup`. It shows the
// first topic of each note as a preview and adapts to terminal width.
func Startup(w io.Writer, o Options, notes []*note.Note) {
	t := o.Theme()
	if len(notes) == 0 {
		return
	}

	inner := o.width() - 4
	var body strings.Builder

	head := "STICK" + t.Dim.Render("  ·  "+itoa(len(notes))+" active")
	body.WriteString(t.Title.Render(head))
	body.WriteString("\n")
	body.WriteString(t.Dim.Render(strings.Repeat("─", min(inner, 40))))
	body.WriteString("\n")

	for _, n := range notes {
		pin := "  "
		if n.IsPinned {
			pin = t.PinMark + " "
		}
		title := truncate(n.Title, inner-4)
		fmt.Fprintf(&body, "%s%s\n", pin, title)
		if len(n.Topics) > 0 {
			fmt.Fprintf(&body, "%s\n", t.Dim.Render(truncate("· "+n.Topics[0].Text, inner-2)))
		}
	}

	foot := "stick list to see everything"
	body.WriteString("\n")
	body.WriteString(t.Dim.Render(foot))

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Padding(0, 1).
		Render(strings.TrimRight(body.String(), "\n"))
	fmt.Fprintln(w, box)
}
