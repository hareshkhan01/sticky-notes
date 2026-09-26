package note

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Service implements all note business rules on top of a Repository.
type Service struct {
	repo Repository
	now  func() time.Time // replaceable in tests
}

// NewService wires a service to a repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// Close releases the underlying repository resources.
func (s *Service) Close() error { return s.repo.Close() }

// Create stores a new note. raw holds one topic per line; an explicit
// title wins over the auto-derived one.
func (s *Service) Create(ctx context.Context, title, raw string) (*Note, error) {
	candidates := Sanitize(raw)
	if len(candidates) == 0 {
		return nil, ErrEmpty
	}
	if len(candidates) > MaxTopics {
		return nil, ErrTooManyTopics
	}
	topics := make([]Topic, len(candidates))
	for i, c := range candidates {
		text, err := ValidateTopic(c)
		if err != nil {
			return nil, err
		}
		topics[i] = Topic{Position: i, Text: text}
	}
	name, err := ValidateTitle(title)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = DeriveTitle(topics)
	}
	now := s.now().UTC()
	n := &Note{Title: name, Topics: topics, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// Get returns one note with topics in display order.
func (s *Service) Get(ctx context.Context, id int64) (*Note, error) {
	return s.get(ctx, id)
}

// List returns notes ordered by updated_at (newest first) with pinned
// notes hoisted to the front when pinnedFirst is set.
func (s *Service) List(ctx context.Context, f NoteFilter) ([]*Note, error) {
	return s.repo.List(ctx, f)
}

// Search matches query against titles and topic text.
func (s *Service) Search(ctx context.Context, query string, f NoteFilter) ([]*Note, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, ErrEmpty
	}
	return s.repo.Search(ctx, q, f)
}

// EditTitle renames a note.
func (s *Service) EditTitle(ctx context.Context, id int64, title string) error {
	n, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	name, err := ValidateTitle(title)
	if err != nil {
		return err
	}
	if name == "" {
		return ErrEmpty
	}
	n.Title = name
	n.UpdatedAt = s.now().UTC()
	return s.repo.Update(ctx, n)
}

// ReplaceTopics swaps a note's topics for a new set. An empty set clears
// the topics but keeps the note itself.
func (s *Service) ReplaceTopics(ctx context.Context, id int64, raw string) error {
	n, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	candidates := Sanitize(raw)
	if len(candidates) > MaxTopics {
		return ErrTooManyTopics
	}
	topics := make([]Topic, 0, len(candidates))
	for i, c := range candidates {
		text, err := ValidateTopic(c)
		if err != nil {
			return err
		}
		topics = append(topics, Topic{Position: i, Text: text})
	}
	n.Topics = topics
	if n.Title == "" && len(topics) > 0 {
		n.Title = DeriveTitle(topics)
	}
	n.UpdatedAt = s.now().UTC()
	return s.repo.Update(ctx, n)
}

// Delete removes a note and its topics.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.get(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// SetPin and SetArchive toggle note state, bumping updated_at.
func (s *Service) SetPin(ctx context.Context, id int64, pinned bool) error {
	return s.setFlags(ctx, id, func(n *Note) { n.IsPinned = pinned })
}

// SetArchive toggles the archived flag on a note.
func (s *Service) SetArchive(ctx context.Context, id int64, archived bool) error {
	return s.setFlags(ctx, id, func(n *Note) { n.IsArchived = archived })
}

func (s *Service) setFlags(ctx context.Context, id int64, apply func(*Note)) error {
	n, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	apply(n)
	return s.repo.SetFlags(ctx, id, n.IsPinned, n.IsArchived, s.now().UTC())
}

func (s *Service) get(ctx context.Context, id int64) (*Note, error) {
	n, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, NotFoundError{ID: id}
		}
		return nil, err
	}
	return n, nil
}
