package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hareshkhan01/sticky-notes/internal/note"
	"github.com/hareshkhan01/sticky-notes/internal/ui"
)

var editFlags struct {
	title     string
	editor    bool
	topics    []string
	addTopics []string
}

// newEditCmd builds `stick edit`.
func newEditCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a note's title or topics",
		Long: `Edit an existing note.

With no flags the note is opened in $EDITOR (or $VISUAL) as plain text:
line 1 is the title, every following non-empty line is one topic, with a
maximum of 5 topics and 200 characters each.

Use --title, --topic, or --add for quick changes without an editor.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(cmd, args[0])
		},
	}
	c.Flags().StringVarP(&editFlags.title, "title", "t", "", "replace the note title")
	c.Flags().StringSliceVar(&editFlags.topics, "topic", nil, "replace all topics (comma separated)")
	c.Flags().StringSliceVar(&editFlags.addTopics, "add", nil, "append topics (comma separated)")
	c.Flags().BoolVar(&editFlags.editor, "editor", false, "force the $EDITOR flow")
	return c
}

func runEdit(cmd *cobra.Command, arg string) error {
	c := newCLI(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
	if err := c.setup(); err != nil {
		return err
	}
	defer c.close()
	if err := c.open(); err != nil {
		return err
	}
	id, err := resolveID(arg)
	if err != nil {
		return err
	}
	n, err := c.svc.Get(cmd.Context(), id)
	if err != nil {
		return err
	}

	// Editor flow is the default when no quick flags were supplied.
	if editFlags.editor || (editFlags.title == "" && editFlags.topics == nil && editFlags.addTopics == nil) {
		return editInEditor(cmd, c, n)
	}

	changed := false
	if editFlags.title != "" {
		if err := c.svc.EditTitle(cmd.Context(), id, editFlags.title); err != nil {
			return err
		}
		changed = true
	}
	if editFlags.topics != nil {
		if err := c.svc.ReplaceTopics(cmd.Context(), id, strings.Join(editFlags.topics, "\n")); err != nil {
			return err
		}
		changed = true
	}
	if editFlags.addTopics != nil {
		merged := currentTopicText(n)
		for _, t := range editFlags.addTopics {
			merged = append(merged, t)
		}
		if err := c.svc.ReplaceTopics(cmd.Context(), id, strings.Join(merged, "\n")); err != nil {
			return err
		}
		changed = true
	}

	if changed {
		updated, err := c.svc.Get(cmd.Context(), id)
		if err != nil {
			return err
		}
		fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Note updated"))
		ui.NotePage(c.stdout, c.ui, updated)
	} else {
		fmt.Fprintln(c.stdout, "Nothing changed.")
	}
	return nil
}

// currentTopicText extracts topic lines in display order.
func currentTopicText(n *note.Note) []string {
	out := make([]string, len(n.Topics))
	for i, t := range n.Topics {
		out[i] = t.Text
	}
	return out
}

// editInEditor dumps the note to a temp file, opens $EDITOR, and applies
// whatever came back. An unchanged buffer leaves the note untouched.
func editInEditor(cmd *cobra.Command, c *cli, n *note.Note) error {
	editor := firstNonEmpty(os.Getenv("VISUAL"), os.Getenv("EDITOR"))
	if editor == "" {
		return note.ErrNoEditor
	}

	before := serializeNote(n)
	tmp, err := os.CreateTemp("", "stick-note-*.md")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.WriteString(before); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	ed := exec.Command(editor, tmpPath)
	ed.Stdin = c.stdin
	ed.Stdout = c.stdout
	ed.Stderr = c.stderr
	ed.Dir = filepath.Dir(tmpPath)
	if err := ed.Run(); err != nil {
		return fmt.Errorf("editor %q failed: %w", editor, err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("read temp file: %w", err)
	}
	after := string(data)
	if after == before {
		fmt.Fprintln(c.stdout, "No changes; note left as is.")
		return nil
	}

	title, body := splitEditedNote(after)
	if title != n.Title {
		if err := c.svc.EditTitle(cmd.Context(), n.ID, title); err != nil {
			return err
		}
	}
	if err := c.svc.ReplaceTopics(cmd.Context(), n.ID, body); err != nil {
		return err
	}

	updated, err := c.svc.Get(cmd.Context(), n.ID)
	if err != nil {
		return err
	}
	fmt.Fprintln(c.stdout, c.ui.Theme().Ok.Render("✓ Note updated"))
	ui.NotePage(c.stdout, c.ui, updated)
	return nil
}

// serializeNote renders the editable plain-text form: title on the first
// line, a blank separator, then one topic per line.
func serializeNote(n *note.Note) string {
	var b strings.Builder
	b.WriteString(n.Title)
	b.WriteString("\n\n")
	for _, t := range n.Topics {
		b.WriteString(t.Text)
		b.WriteString("\n")
	}
	return b.String()
}

// splitEditedNote parses the edited buffer back into title and body.
func splitEditedNote(s string) (title, body string) {
	lines := strings.SplitN(s, "\n", 2)
	title = strings.TrimSpace(lines[0])
	if len(lines) == 2 {
		body = lines[1]
	}
	return title, body
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
