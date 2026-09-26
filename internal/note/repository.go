package note

import (
	"context"
	"time"
)

// NoteFilter selects which notes a listing should return.
type NoteFilter struct {
	IncludeArchived bool
	PinnedFirst     bool
	Limit           int // <= 0 means no limit
}

// Repository is the persistence contract the service depends on. The
// SQLite implementation lives in internal/storage; tests substitute a fake.
type Repository interface {
	Create(ctx context.Context, n *Note) error
	Update(ctx context.Context, n *Note) error
	Delete(ctx context.Context, id int64) error
	// SetFlags pins/unpins or archives/unarchives in one write, updating
	// updatedAt to the supplied instant.
	SetFlags(ctx context.Context, id int64, pinned, archived bool, updatedAt time.Time) error
	Get(ctx context.Context, id int64) (*Note, error)
	List(ctx context.Context, f NoteFilter) ([]*Note, error)
	Search(ctx context.Context, query string, f NoteFilter) ([]*Note, error)
	Close() error
}
