// Package note contains the sticky-note domain model, validation rules,
// and the business services that operate on notes. It deliberately knows
// nothing about SQL, Cobra, or terminal rendering.
package note

import "time"

// Limits applied to every note. They are constants rather than user
// configuration so validation rules stay consistent across releases.
const (
	// MaxTopics is the maximum number of topics a single note may hold.
	MaxTopics = 5
	// MaxTopicRunes is the character budget for a single topic.
	MaxTopicRunes = 200
	// MaxTitleRunes is the character budget for a note title.
	MaxTitleRunes = 60
)

// Topic is a single bullet on a note page.
type Topic struct {
	Position int    // 0-based, defines display order
	Text     string // at most MaxTopicRunes runes
}

// Note is a sticky-note page: a title plus up to MaxTopics topics.
type Note struct {
	ID         int64
	Title      string
	Topics     []Topic
	IsPinned   bool
	IsArchived bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsEmpty reports whether the note carries no title and no topics.
func (n *Note) IsEmpty() bool {
	return n.Title == "" && len(n.Topics) == 0
}
