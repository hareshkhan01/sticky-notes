package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hareshkhan01/sticky-notes/internal/note"
)

// openTempStore creates a Store backed by a fresh database in t.TempDir.
func openTempStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "notes.db")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestOpenMigrationsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.db")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	// Re-opening the same file must succeed and not duplicate schema.
	st2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	st2.Close()
}

func TestStoreCRUDRoundTrip(t *testing.T) {
	st := openTempStore(t)
	ctx := context.Background()

	n := &note.Note{
		Title: "errands",
		Topics: []note.Topic{
			{Position: 0, Text: "buy milk"},
			{Position: 1, Text: "post letter"},
		},
	}
	if err := st.Create(ctx, n); err != nil {
		t.Fatal(err)
	}
	if n.ID == 0 {
		t.Fatal("expected ID to be assigned")
	}

	got, err := st.Get(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "errands" || len(got.Topics) != 2 || got.Topics[1].Text != "post letter" {
		t.Fatalf("round trip mismatch: %+v", got)
	}

	got.Topics = got.Topics[:1]
	got.Title = "errands v2"
	if err := st.Update(ctx, got); err != nil {
		t.Fatal(err)
	}
	again, _ := st.Get(ctx, n.ID)
	if again.Title != "errands v2" || len(again.Topics) != 1 {
		t.Fatalf("update mismatch: %+v", again)
	}

	if err := st.Delete(ctx, n.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Get(ctx, n.ID); err != note.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestStoreDeleteCascadesTopics(t *testing.T) {
	st := openTempStore(t)
	ctx := context.Background()

	n := &note.Note{Title: "t", Topics: []note.Topic{{Position: 0, Text: "x"}}}
	if err := st.Create(ctx, n); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM topics`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 topic row, got %d", count)
	}
	if err := st.Delete(ctx, n.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM topics`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected topics cascade-deleted, got %d rows", count)
	}
}

func TestStoreListPinnedFirstAndLimit(t *testing.T) {
	st := openTempStore(t)
	ctx := context.Background()

	for i, title := range []string{"a", "b", "c"} {
		n := &note.Note{Title: title, Topics: []note.Topic{{Position: 0, Text: title}}}
		if err := st.Create(ctx, n); err != nil {
			t.Fatal(err)
		}
		if i == 2 {
			if err := st.SetFlags(ctx, n.ID, true, false, n.UpdatedAt); err != nil {
				t.Fatal(err)
			}
		}
	}

	notes, err := st.List(ctx, note.NoteFilter{PinnedFirst: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 3 || notes[0].Title != "c" {
		t.Fatalf("expected pinned note first, got %+v", notes)
	}

	limited, err := st.List(ctx, note.NoteFilter{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(limited) != 2 {
		t.Fatalf("expected limit honoured, got %d", len(limited))
	}
}

func TestStoreSearchMatchesTopicText(t *testing.T) {
	st := openTempStore(t)
	ctx := context.Background()

	n1 := &note.Note{Title: "work", Topics: []note.Topic{{Position: 0, Text: "fix the login bug"}}}
	n2 := &note.Note{Title: "home", Topics: []note.Topic{{Position: 0, Text: "water the plants"}}}
	for _, n := range []*note.Note{n1, n2} {
		if err := st.Create(ctx, n); err != nil {
			t.Fatal(err)
		}
	}

	hits, err := st.Search(ctx, "login", note.NoteFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Title != "work" {
		t.Fatalf("expected only 'work' to match, got %+v", hits)
	}

	// Wildcards in user input are matched literally.
	none, err := st.Search(ctx, "100%", note.NoteFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Fatalf("wildcards should not match everything, got %d", len(none))
	}
}

func TestStoreCountActiveIgnoresArchived(t *testing.T) {
	st := openTempStore(t)
	ctx := context.Background()

	n1 := &note.Note{Title: "one", Topics: []note.Topic{{Position: 0, Text: "x"}}}
	n2 := &note.Note{Title: "two", Topics: []note.Topic{{Position: 0, Text: "y"}}}
	for _, n := range []*note.Note{n1, n2} {
		if err := st.Create(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.SetFlags(ctx, n2.ID, false, true, n2.UpdatedAt); err != nil {
		t.Fatal(err)
	}

	got, err := st.CountActive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("expected 1 active note, got %d", got)
	}

	archived, err := st.List(ctx, note.NoteFilter{IncludeArchived: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(archived) != 2 {
		t.Fatalf("expected --all to see 2 notes, got %d", len(archived))
	}
}
