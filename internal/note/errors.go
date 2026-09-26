package note

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by the service. Callers should compare with
// errors.Is rather than matching on strings.
var (
	// ErrNotFound means no note exists with the requested ID.
	ErrNotFound = errors.New("note not found")
	// ErrEmpty is returned when note text contains no visible content.
	ErrEmpty = errors.New("note text is empty")
	// ErrTooManyTopics means the caller tried to add more than MaxTopics.
	ErrTooManyTopics = errors.New("a note can hold at most 5 topics")
	// ErrNoEditor is returned when interactive editing is impossible.
	ErrNoEditor = errors.New("no usable editor; set $EDITOR or use flags")
)

// ErrTitleTooLong reports a title exceeding the rune budget.
type ErrTitleTooLong struct{ Length int }

func (e ErrTitleTooLong) Error() string {
	return fmt.Sprintf("title is %d characters, the limit is %d", e.Length, MaxTitleRunes)
}

// ErrTopicTooLong reports a topic exceeding the rune budget.
type ErrTopicTooLong struct {
	Index  int // 0-based position of the offending topic
	Length int
}

func (e ErrTopicTooLong) Error() string {
	return fmt.Sprintf("topic %d is %d characters, the limit is %d", e.Index+1, e.Length, MaxTopicRunes)
}

// NotFoundError reports a specific missing note ID.
type NotFoundError struct{ ID int64 }

func (e NotFoundError) Error() string {
	return fmt.Sprintf("note %d not found", e.ID)
}

// IsNotFound reports whether err represents a missing note.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
