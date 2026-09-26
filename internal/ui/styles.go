// Package ui renders notes as terminal output. It takes domain types and
// returns strings; it never talks to storage or interprets commands.
package ui

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
	"github.com/rivo/uniseg"
)

// maxContentWidth keeps every box readable inside an 80-column terminal.
const maxContentWidth = 76

// Options controls rendering for one invocation of the CLI.
type Options struct {
	NoColor    bool   // --no-color flag
	JSON       bool   // --json flag
	ColorMode  string // "auto", "always", "never" from config
	DateFormat string // Go reference layout from config
	Width      int    // terminal width; <=0 means detect
}

// colorEnabled applies the precedence chain: flag beats environment beats
// config; piped output is never styled.
func (o Options) colorEnabled() bool {
	if o.NoColor {
		return false
	}
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}
	switch o.ColorMode {
	case "never":
		return false
	case "always":
		return true
	}
	return isatty.IsTerminal(os.Stdout.Fd())
}

// width returns the usable content width for boxes and tables.
func (o Options) width() int {
	w := o.Width
	if w <= 0 {
		if _, cols, err := term.GetSize(os.Stdout.Fd()); err == nil && cols > 0 {
			w = cols
		} else {
			w = 80
		}
	}
	if w > maxContentWidth {
		w = maxContentWidth
	}
	if w < 24 {
		w = 24
	}
	return w
}

// theme holds every lipgloss style in one place. When color is off the
// pin mark degrades to a plain asterisk and styles become no-ops.
type theme struct {
	Title   lipgloss.Style
	Dim     lipgloss.Style
	Ok      lipgloss.Style
	Warn    lipgloss.Style
	Border  lipgloss.AdaptiveColor
	PinMark string
}

func (o Options) Theme() theme {
	if !o.colorEnabled() {
		return theme{PinMark: "*"}
	}
	ink := lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
	muted := lipgloss.AdaptiveColor{Light: "243", Dark: "244"}
	accent := lipgloss.AdaptiveColor{Light: "136", Dark: "179"} // muted amber

	return theme{
		Title:   lipgloss.NewStyle().Foreground(ink).Bold(true),
		Dim:     lipgloss.NewStyle().Foreground(muted),
		Ok:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "108"}),
		Warn:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "173"}),
		Border:  accent,
		PinMark: "📌",
	}
}

// truncate cuts s to at most max display cells, appending an ellipsis when
// the string is actually shortened. Grapheme-aware so emoji and CJK do
// not break table alignment.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if uniseg.StringWidth(s) <= max {
		return s
	}
	budget := max - 1 // reserve one cell for the ellipsis
	var b strings.Builder
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		w := gr.Width()
		if w > budget {
			break
		}
		b.WriteString(gr.Str())
		budget -= w
	}
	return b.String() + "…"
}

// pad right-pads s with spaces to exactly width display cells.
func pad(s string, width int) string {
	d := width - uniseg.StringWidth(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}

// itoa is a small convenience wrapper.
func itoa(n int) string { return strconv.Itoa(n) }
