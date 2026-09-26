package cmd

import "github.com/hareshkhan01/sticky-notes/internal/note"

// noteJSON is the JSON shape for a single note. Topics are plain strings
// in display order.
type noteJSON struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Topics     []string `json:"topics"`
	IsPinned   bool     `json:"is_pinned"`
	IsArchived bool     `json:"is_archived"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

// noteToJSON converts a domain note to its JSON representation. The
// topics slice is never nil so JSON always contains an array.
func noteToJSON(n *note.Note) noteJSON {
	topics := make([]string, len(n.Topics))
	for i, t := range n.Topics {
		topics[i] = t.Text
	}
	return noteJSON{
		ID:         n.ID,
		Title:      n.Title,
		Topics:     topics,
		IsPinned:   n.IsPinned,
		IsArchived: n.IsArchived,
		CreatedAt:  n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  n.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// notesToJSON converts a slice of notes.
func notesToJSON(notes []*note.Note) []noteJSON {
	out := make([]noteJSON, 0, len(notes))
	for _, n := range notes {
		out = append(out, noteToJSON(n))
	}
	return out
}
