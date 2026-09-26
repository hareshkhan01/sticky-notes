-- migration 0001: initial schema

CREATE TABLE IF NOT EXISTS notes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT    NOT NULL DEFAULT '',
    is_pinned   INTEGER NOT NULL DEFAULT 0,
    is_archived INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT    NOT NULL,
    updated_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS topics (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id  INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 4),
    text     TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_topics_note ON topics(note_id, position);

CREATE INDEX IF NOT EXISTS idx_notes_active_updated
    ON notes(is_archived, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_notes_pinned
    ON notes(is_pinned, is_archived);
