# Stick — Design Thoughts

Working notes made before and during development. This file records *why*
things are the way they are, so future changes have context.

## 1. Reading of the user's request

The user asked for a terminal sticky-notes app with these specific
requirements, on top of the provided spec (`stick-notes-agent-spec.md`):

1. **Go standard project structure** — `cmd/`, `internal/`, `main.go`.
2. **Proper Go naming conventions** — files `snake_case.go` (Go actually
   prefers short lowercase names, no underscores where possible), exported
   identifiers `PascalCase`, unexported `camelCase`, acronyms capitalised
   (`ID`, `URL`).
3. **Verify Cobra is usable** — done: `go get github.com/spf13/cobra`
   resolves to v1.10.2 in this environment. Confirmed before writing code.
4. **Startup display** — when the user opens a terminal, the latest/pinned
   notes are shown. Opt-in via shell hook, never automatic.
5. **Unlimited notes, but a "page" limit** — a user can create as many
   notes as they want, but *each note* holds at most **5 topics**, and each
   topic's text is length-restricted. This is the main deviation from the
   stock spec: the spec's `Note` model is title+content; ours is
   title + up to 5 topics.
6. **Should not look "AI/vibe coded"** — restrained visual design: no
   rainbow colors, no emoji spam, consistent borders, plain-mode support.
   Also: no leftover TODO litter, no dead code, small focused files.

## 2. Data model decision

A note is a *sticky note page*:

```
Note
├── ID, Title, IsPinned, IsArchived, CreatedAt, UpdatedAt
└── Topics[5]   (ordered, each ≤ MaxTopicRunes)
```

- **Why 5 topics per note, hard limit?** The user asked for it. It also
  matches the real-world metaphor: a physical sticky note holds a handful
  of bullet points, not a document. The limit is enforced in the service
  layer (`note.MaxTopics = 5`), not in the UI, so every path (CLI, future
  TUI, import) hits the same rule.
- **Why a separate `topics` table instead of a JSON column?** SQL queries
  (search, ordered retrieval) stay simple and indexable; a JSON blob would
  push ordering/matching logic into Go and make `search` scan everything.
  The `position` column (0–4) preserves topic order.
- **Character limit per topic: 200 runes.** "Characters" is interpreted as
  Unicode runes, not bytes, so emoji/CJK count fairly. Enforced centrally
  in `note.ValidateTitle` / `note.ValidateTopic`. Configurable in code as a
  named constant, not user config — changing it silently would orphan old
  validation rules.
- **Title: 60 runes**, so list-table columns never need to guess.
- Titles are required for `add --title`, auto-derived from the first topic
  otherwise (truncated to 60 runes). This keeps `stick add "text"` a
  one-shot capture.

## 3. Storage

- **`modernc.org/sqlite` (pure Go, no CGO)** instead of `mattn/go-sqlite3`.
  Reason: CGO makes cross-compilation painful and breaks `go build` on
  machines without a C toolchain. modernc is a maintained, widely used
  transpiler of SQLite to Go. Slight perf cost is irrelevant at this scale.
- One `*sql.DB` per process, `WAL` journal mode, `busy_timeout=5s`,
  `foreign_keys=ON`. WAL lets a `stick startup` read never block on a
  stale write lock.
- Schema versioning via a `schema_version` table and embedded migration
  files (`storage/migrations/*.sql`, `embed.FS`). Migration 1 creates
  `notes` and `topics`. Each migration runs in a transaction.
- Timestamps stored as RFC 3339 strings in UTC — human-inspectable with
  `sqlite3`, sorts correctly, and `time.Time` round-trips losslessly.

## 4. Architecture

Strict one-way dependency flow (spec §7):

```
cmd (Cobra) → note.Service → note.Repository (interface) → storage.SQLite
                        ↓
                      ui (rendering only)
```

- `note.Repository` is defined in the **note** package, not storage — the
  consumer owns the interface (Go idiom), and it makes the service testable
  with an in-memory fake without importing SQLite.
- Cobra command files contain **no SQL and no rendering logic**; they parse
  flags, call the service, and hand results to `ui`.
- `ui` never touches the database. It takes domain types and prints.

## 5. Rendering choices (the "not vibe-coded" part)

- **Lip Gloss** for borders/colors, with a central `styles.go`. One accent
  color (a muted yellow, fitting for sticky notes), one dim gray for
  metadata. No gradients, no random emoji, no full-width art.
- Pin marker is `*` in plain mode and `📌` only in color mode. Emoji width
  is the classic way these tables end up misaligned, so `ui` measures with
  `runewidth`-aware truncation everywhere it cuts a string.
- Every render path honours, in order: `--no-color` flag → `NO_COLOR` env →
  non-TTY stdout (piped) → config `display.color`. Piped output is always
  plain so `stick list | grep foo` behaves.
- Borders are drawn by Lip Gloss with a fixed inner width, so the box never
  overflows an 80-column terminal; long topics wrap instead of pushing the
  right border out.
- `--json` output is `encoding/json` with sorted field order and no ANSI.

## 6. Startup integration

- `stick startup` is *only* run from a managed block the user explicitly
  installs with `stick init <shell>`. The command itself:
  - prints nothing when there are no active notes (and `show_when_empty`
    is false),
  - prints a one-line warning to **stderr** and exits 0 if the DB is
    unreadable — a broken DB must never block a shell from opening,
  - skips rendering entirely when stdout is not a TTY (non-interactive,
    e.g. `ssh host stick startup`).
- Managed blocks are delimited with `# >>> stick notes >>>` /
  `# <<< stick notes <<<` (fish uses the same markers with fish syntax).
  `init` is idempotent: if the marker exists, it does nothing. `uninstall`
  removes exactly the marker-delimited range and refuses to touch a file
  whose markers are unbalanced (safety valve against manual edits).
- A timestamped backup (`~/.config/stick/backups/`) is written before the
  first modification of a startup file in a given run.
- Interactive-vs-non-interactive is handled *inside* the hook
  (`[[ $- == *i* ]]` for bash/zsh, `status is-interactive` for fish) so
  the binary itself stays fast and does not need to guess.

## 7. Concurrency and CLI behaviour

- Deletion and destructive ops require a TTY confirmation unless
  `--yes` is passed. When stdin is not a TTY and `--yes` is absent, the
  command aborts with a message instead of silently proceeding — this is
  the safer default for scripts.
- Exit codes: 0 success, 1 usage/runtime error, so shell hooks can rely on
  them. `stick startup` never exits non-zero for "no notes".
- Errors are printed once, to stderr, with an actionable hint where
  possible (`note 42 not found — run 'stick list' to see IDs`).

## 8. Testing strategy

- `storage` tests run against temp-dir SQLite files, never the user's real
  DB (`$XDG_DATA_HOME`/`HOME` are pointed at a temp dir in tests).
- `note.Service` tests use a fake repository — no SQL involved — so the
  business rules (topic limits, pinning, search) are tested in isolation.
- Shell install/uninstall tests operate on temp files with fake `$HOME`.
- No test writes to the developer's actual shell configuration.
- Commands are tested by invoking the built Cobra command tree with args
  and capturing output buffers.

## 9. Deliberate non-goals for the MVP

- No Bubble Tea interactive TUI (spec lists it as optional/future).
- No tags, export/import, fuzzy search, project-aware notes (Phase 5).
- No daemon; no network access; no telemetry.
