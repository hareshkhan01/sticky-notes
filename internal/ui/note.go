package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"

	"github.com/hareshkhan01/sticky-notes/internal/note"
)

// NotePage renders a single note with all of its topics, inside a light
// border. Long topics wrap; the box never exceeds the terminal width.
func NotePage(w io.Writer, o Options, n *note.Note) {
	t := o.Theme()
	inner := o.width() - 4 // border + padding on each side

	var body strings.Builder
	pin := "  "
	if n.IsPinned {
		pin = t.PinMark + " "
	}

	head := pin + n.Title
	if n.Title == "" {
		head = pin + "(untitled)"
	}
	body.WriteString(t.Title.Render(truncate(head, inner)))
	body.WriteString("\n")
	body.WriteString(t.Dim.Render(strings.Repeat("─", min(inner, 40))))
	body.WriteString("\n")

	if len(n.Topics) == 0 {
		body.WriteString(t.Dim.Render("(no topics)"))
		body.WriteString("\n")
	} else {
		for i, topic := range n.Topics {
			fmt.Fprintf(&body, " %d. %s\n", i+1, wrap(topic.Text, inner-4))
		}
	}

	body.WriteString("\n")
	body.WriteString(t.Dim.Render(metaLines(o, n)))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1).
		Render(strings.TrimRight(body.String(), "\n"))
	fmt.Fprintln(w, box)
}

// metaLines formats the timestamp/status footer of a note.
func metaLines(o Options, n *note.Note) string {
	layout := o.dateFormat()
	status := "active"
	if n.IsArchived {
		status = "archived"
	}
	if n.IsPinned {
		status += ", pinned"
	}
	return fmt.Sprintf("id %d · updated %s · %s", n.ID, n.UpdatedAt.Format(layout), status)
}

// dateFormat resolves the configured layout with a safe fallback.
func (o Options) dateFormat() string {
	if o.DateFormat == "" {
		return "2006-01-02 15:04"
	}
	return o.DateFormat
}

// wrap breaks s into lines of at most width display cells without
// splitting graphemes. Existing newlines are honoured.
func wrap(s string, width int) string {
	if width < 8 {
		width = 8
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := ""
		for _, word := range words {
			switch {
			case line == "":
				line = word
			case uniseg.StringWidth(line)+1+uniseg.StringWidth(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// Confirm asks a yes/no question on an interactive terminal. It returns
// the parsed answer and whether an answer could be read at all.
func Confirm(r io.Reader, w io.Writer, prompt string) bool {
	fmt.Fprintf(w, "%s [y/N] ", prompt)
	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && line == "" {
		fmt.Fprintln(w)
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// FmtTime formats t using the option's date layout; used by cmd outputs
// so timestamps stay consistent across commands.
func FmtTime(o Options, t time.Time) string {
	return t.Format(o.dateFormat())
}
