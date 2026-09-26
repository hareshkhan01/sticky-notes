package note

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeRepo is an in-memory Repository for service tests.
type fakeRepo struct {
	notes  map[int64]*Note
	topics map[int64][]Topic
	nextID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{notes: map[int64]*Note{}, topics: map[int64][]Topic{}, nextID: 1}
}

func (f *fakeRepo) Create(_ context.Context, n *Note) error {
	n.ID = f.nextID
	f.nextID++
	cp := *n
	f.notes[n.ID] = &cp
	f.topics[n.ID] = append([]Topic(nil), n.Topics...)
	return nil
}

func (f *fakeRepo) Update(_ context.Context, n *Note) error {
	if _, ok := f.notes[n.ID]; !ok {
		return ErrNotFound
	}
	cp := *n
	f.notes[n.ID] = &cp
	f.topics[n.ID] = append([]Topic(nil), n.Topics...)
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.notes[id]; !ok {
		return ErrNotFound
	}
	delete(f.notes, id)
	delete(f.topics, id)
	return nil
}

func (f *fakeRepo) SetFlags(_ context.Context, id int64, pinned, archived bool, updatedAt time.Time) error {
	n, ok := f.notes[id]
	if !ok {
		return ErrNotFound
	}
	n.IsPinned, n.IsArchived, n.UpdatedAt = pinned, archived, updatedAt
	return nil
}

func (f *fakeRepo) Get(_ context.Context, id int64) (*Note, error) {
	n, ok := f.notes[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *n
	cp.Topics = append([]Topic(nil), f.topics[id]...)
	return &cp, nil
}

func (f *fakeRepo) List(_ context.Context, _ NoteFilter) ([]*Note, error) {
	return nil, errors.New("not implemented in fake")
}

func (f *fakeRepo) Search(_ context.Context, _ string, _ NoteFilter) ([]*Note, error) {
	return nil, errors.New("not implemented in fake")
}

func (f *fakeRepo) Close() error { return nil }

func TestServiceCreateLimits(t *testing.T) {
	svc := NewService(newFakeRepo())
	defer svc.Close()

	if _, err := svc.Create(context.Background(), "", "   \n \t "); err == nil {
		t.Fatal("expected ErrEmpty for whitespace-only input")
	}

	six := strings.Repeat("x", 10) + "\n"
	six = strings.Repeat(six, 6)
	if _, err := svc.Create(context.Background(), "", six); !errors.Is(err, ErrTooManyTopics) {
		t.Fatalf("expected ErrTooManyTopics, got %v", err)
	}
}

func TestServiceCreateTopicTooLong(t *testing.T) {
	svc := NewService(newFakeRepo())
	defer svc.Close()

	long := strings.Repeat("あ", MaxTopicRunes+1) // runes, not bytes
	_, err := svc.Create(context.Background(), "", long)
	var tooLong ErrTopicTooLong
	if !errors.As(err, &tooLong) {
		t.Fatalf("expected ErrTopicTooLong, got %v", err)
	}
}

func TestServiceCreateDerivesTitle(t *testing.T) {
	svc := NewService(newFakeRepo())
	defer svc.Close()

	n, err := svc.Create(context.Background(), "", "first topic\nsecond topic")
	if err != nil {
		t.Fatal(err)
	}
	if n.Title != "first topic" {
		t.Fatalf("expected derived title %q, got %q", "first topic", n.Title)
	}
	if len(n.Topics) != 2 || n.Topics[1].Text != "second topic" {
		t.Fatalf("unexpected topics: %+v", n.Topics)
	}
}

func TestServiceSanitizeStripsBullets(t *testing.T) {
	got := Sanitize("- one\n* two\n• three\n\n   four  ")
	want := []string{"one", "two", "three", "four"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("topic %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestServiceSetPinBumpsUpdated(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	defer svc.Close()

	n, err := svc.Create(context.Background(), "t", "body")
	if err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Hour)
	svc.now = func() time.Time { return later }
	if err := svc.SetPin(context.Background(), n.ID, true); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(context.Background(), n.ID)
	if !got.IsPinned || !got.UpdatedAt.Equal(later.UTC()) {
		t.Fatalf("pin not applied correctly: %+v", got)
	}
}

func TestServiceNotFoundWrapsID(t *testing.T) {
	svc := NewService(newFakeRepo())
	defer svc.Close()

	err := svc.Delete(context.Background(), 99)
	var nf NotFoundError
	if !errors.As(err, &nf) || nf.ID != 99 {
		t.Fatalf("expected NotFoundError{99}, got %v", err)
	}
}
