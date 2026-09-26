// Package storage provides the SQLite-backed implementation of the note
// repository, including schema migrations and connection management.
package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hareshkhan01/sticky-notes/internal/note"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// schemaVersionTable is created before any migration runs so the version
// stamp itself survives an interrupted first boot.
const schemaVersionTable = `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL
);`

// Store is a SQLite-backed note.Repository.
type Store struct {
	db *sql.DB
}

// Open creates the parent directory (mode 0o700 where supported), opens
// the database, applies pending migrations and returns a ready Store.
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage: database path is empty")
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("storage: create database directory: %w", err)
		}
	}
	// WAL keeps concurrent readers cheap; busy_timeout avoids spurious
	// "database is locked" errors when a previous process is mid-write.
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open database: %w", err)
	}
	// modernc's driver serialises writes; one connection keeps things
	// simple and avoids SQLITE_BUSY under parallel CLI invocations.
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// Ping verifies the database is reachable.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

// Migrate applies every pending migration, newest first by filename.
func (s *Store) migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin migration: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, schemaVersionTable); err != nil {
		return fmt.Errorf("storage: create schema_version: %w", err)
	}
	var version int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return fmt.Errorf("storage: read schema version: %w", err)
	}

	entries, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("storage: list migrations: %w", err)
	}
	sort.Strings(entries)

	for _, name := range entries {
		v, err := migrationVersion(name)
		if err != nil {
			return err
		}
		if v <= version {
			continue
		}
		body, err := migrationFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("storage: read migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("storage: apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_version (version) VALUES (?)`, v); err != nil {
			return fmt.Errorf("storage: stamp migration %s: %w", name, err)
		}
	}
	return tx.Commit()
}

func migrationVersion(name string) (int, error) {
	base := filepath.Base(name)
	var v int
	if _, err := fmt.Sscanf(base, "%d", &v); err != nil {
		return 0, fmt.Errorf("storage: migration %s lacks numeric prefix", name)
	}
	return v, nil
}

// layout is the RFC 3339 second-precision format used for all timestamps.
const layout = time.RFC3339

func formatTime(t time.Time) string { return t.UTC().Format(layout) }

func parseTime(s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ---- note.Repository ----

// Create persists a new note and its topics, assigning the note ID.
func (s *Store) Create(ctx context.Context, n *note.Note) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin create: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO notes (title, is_pinned, is_archived, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		n.Title, boolToInt(n.IsPinned), boolToInt(n.IsArchived), formatTime(n.CreatedAt), formatTime(n.UpdatedAt))
	if err != nil {
		return fmt.Errorf("storage: insert note: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("storage: note id: %w", err)
	}
	if err := insertTopics(ctx, tx, id, n.Topics); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: commit create: %w", err)
	}
	n.ID = id
	return nil
}

// Update rewrites title and topics for an existing note.
func (s *Store) Update(ctx context.Context, n *note.Note) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin update: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE notes SET title = ?, updated_at = ? WHERE id = ?`,
		n.Title, formatTime(n.UpdatedAt), n.ID)
	if err != nil {
		return fmt.Errorf("storage: update note: %w", err)
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return note.NotFoundError{ID: n.ID}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM topics WHERE note_id = ?`, n.ID); err != nil {
		return fmt.Errorf("storage: clear topics: %w", err)
	}
	if err := insertTopics(ctx, tx, n.ID, n.Topics); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: commit update: %w", err)
	}
	return nil
}

func insertTopics(ctx context.Context, tx *sql.Tx, noteID int64, topics []note.Topic) error {
	for _, t := range topics {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO topics (note_id, position, text) VALUES (?, ?, ?)`,
			noteID, t.Position, t.Text); err != nil {
			return fmt.Errorf("storage: insert topic: %w", err)
		}
	}
	return nil
}

// Delete removes a note; topics go with it via ON DELETE CASCADE.
func (s *Store) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete note: %w", err)
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return note.NotFoundError{ID: id}
	}
	return nil
}

// SetFlags pins/unpins or archives/unarchives in a single write.
func (s *Store) SetFlags(ctx context.Context, id int64, pinned, archived bool, updatedAt time.Time) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE notes SET is_pinned = ?, is_archived = ?, updated_at = ? WHERE id = ?`,
		boolToInt(pinned), boolToInt(archived), formatTime(updatedTime(updatedAt)), id)
	if err != nil {
		return fmt.Errorf("storage: update flags: %w", err)
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return note.NotFoundError{ID: id}
	}
	return nil
}

// updatedTime guards against a zero time slipping into storage.
func updatedTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

// Get loads a single note with topics ordered by position.
func (s *Store) Get(ctx context.Context, id int64) (*note.Note, error) {
	n := &note.Note{}
	var pinned, archived int
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, is_pinned, is_archived, created_at, updated_at FROM notes WHERE id = ?`, id).
		Scan(&n.ID, &n.Title, &pinned, &archived, &createdAt, &updatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, note.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("storage: get note: %w", err)
	}
	n.IsPinned, n.IsArchived = pinned == 1, archived == 1
	n.CreatedAt, n.UpdatedAt = parseTime(createdAt), parseTime(updatedAt)

	topics, err := s.loadTopics(ctx, id)
	if err != nil {
		return nil, err
	}
	n.Topics = topics
	return n, nil
}

func (s *Store) loadTopics(ctx context.Context, noteID int64) ([]note.Topic, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT position, text FROM topics WHERE note_id = ? ORDER BY position`, noteID)
	if err != nil {
		return nil, fmt.Errorf("storage: load topics: %w", err)
	}
	defer rows.Close()

	var topics []note.Topic
	for rows.Next() {
		var t note.Topic
		if err := rows.Scan(&t.Position, &t.Text); err != nil {
			return nil, fmt.Errorf("storage: scan topic: %w", err)
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

// CountActive returns the number of non-archived notes.
func (s *Store) CountActive(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notes WHERE is_archived = 0`).Scan(&n); err != nil {
		return 0, fmt.Errorf("storage: count notes: %w", err)
	}
	return n, nil
}

// List returns notes ordered newest-first, with pinned notes first when
// the filter asks for it.
func (s *Store) List(ctx context.Context, f note.NoteFilter) ([]*note.Note, error) {
	order := "updated_at DESC"
	if f.PinnedFirst {
		order = "is_pinned DESC, updated_at DESC"
	}
	q := `SELECT id, title, is_pinned, is_archived, created_at, updated_at FROM notes
	      WHERE (? = 1 OR is_archived = 0) ORDER BY ` + order
	args := []any{boolToInt(f.IncludeArchived)}
	if f.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, f.Limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list notes: %w", err)
	}
	defer rows.Close()
	return scanNotes(ctx, rows, s)
}

// Search matches the query against note titles and topic text using
// parameterized LIKE, so user input can never alter the statement.
func (s *Store) Search(ctx context.Context, query string, f note.NoteFilter) ([]*note.Note, error) {
	pattern := "%" + escapeLike(strings.TrimSpace(query)) + "%"
	order := "updated_at DESC"
	if f.PinnedFirst {
		order = "is_pinned DESC, updated_at DESC"
	}
	q := `SELECT DISTINCT n.id, n.title, n.is_pinned, n.is_archived, n.created_at, n.updated_at
	      FROM notes n
	      LEFT JOIN topics t ON t.note_id = n.id
	      WHERE (? = 1 OR n.is_archived = 0)
	        AND (n.title LIKE ? ESCAPE '\' OR t.text LIKE ? ESCAPE '\')
	      ORDER BY ` + order
	args := []any{boolToInt(f.IncludeArchived), pattern, pattern}
	if f.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, f.Limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: search notes: %w", err)
	}
	defer rows.Close()
	return scanNotes(ctx, rows, s)
}

func scanNotes(ctx context.Context, rows *sql.Rows, s *Store) ([]*note.Note, error) {
	// Drain and close the notes cursor before issuing any per-note topic
	// query: the pool holds a single connection (see Open), so a second
	// query while rows is still open would deadlock.
	var out []*note.Note
	for rows.Next() {
		n := &note.Note{}
		var pinned, archived int
		var createdAt, updatedAt string
		if err := rows.Scan(&n.ID, &n.Title, &pinned, &archived, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("storage: scan note: %w", err)
		}
		n.IsPinned, n.IsArchived = pinned == 1, archived == 1
		n.CreatedAt, n.UpdatedAt = parseTime(createdAt), parseTime(updatedAt)
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: scan notes: %w", err)
	}
	rows.Close()

	for _, n := range out {
		topics, err := s.loadTopics(ctx, n.ID)
		if err != nil {
			return nil, err
		}
		n.Topics = topics
	}
	return out, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// escapeLike neutralises LIKE wildcards so user input is matched
// literally and can never alter the query structure.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
