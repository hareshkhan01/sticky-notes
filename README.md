# Stick

**Sticky notes for your terminal.**

Stick is a lightweight CLI application that lets you capture, organize, and manage sticky notes directly from your terminal.

Each note is a page holding up to 5 topics, with a 200-character limit per topic.

Notes are stored in a local SQLite database. Nothing leaves your machine.

**Demo**
<img width="2285" height="1165" alt="demo" src="https://github.com/user-attachments/assets/666f2f80-df5e-47b1-9aaf-3c514406e2b6" />


## Features

- **Quick Capture** — Create notes instantly from your terminal.
- **Organized Notes** — Group up to 5 topics into a single note.
- **Pinned Notes** — Keep important notes at the top.
- **Search** — Find notes by title or topic.
- **Archive** — Hide notes without permanently deleting them.
- **Shell Startup Panel** — Display your notes when opening a terminal.
- **JSON Output** — Get machine-readable output for scripts and automation.
- **Local Storage** — SQLite database with no cloud dependency.
- **Cross-Platform** — Supports Linux and macOS.

## Installation

### Quick Install (Recommended)

Install Stick with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/hareshkhan01/sticky-notes/main/install.sh | sh
```

The installer automatically detects your operating system and CPU architecture, downloads the appropriate precompiled binary, and installs it.

No Go installation or manual compilation is required.

### Supported Platforms

| Operating System | Architecture |
|------------------|--------------|
| Linux | AMD64 (x86_64) |
| Linux | ARM64 (aarch64) |
| macOS | AMD64 (Intel) |
| macOS | ARM64 (Apple Silicon) |

### Installation Directory

The installer places the binary in:

```bash
~/.local/bin/stick
```

If `~/.local/bin` is not in your `PATH`, add it to your shell configuration:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Restart your terminal or reload your shell configuration.

Verify the installation:

```bash
stick version
```

### Download a Binary Manually

You can also download a precompiled binary from GitHub Releases.

**[Download the latest release](https://github.com/hareshkhan01/sticky-notes/releases/latest)**

Choose the archive matching your operating system and architecture, extract it, and place the `stick` binary somewhere in your `PATH`.

### Install from Source

If you have Go installed, you can install Stick directly:

```bash
go install github.com/hareshkhan01/sticky-notes@latest
```

Alternatively, build and install from source:

```bash
git clone https://github.com/hareshkhan01/sticky-notes.git
cd sticky-notes
make install
```

Requires Go.

---

## Quick Start

### Create a Note

```bash
stick "remember the milk"
```

### Create a Note with a Title

```bash
stick add --title "Errands" "buy milk
post letter"
```

### List Notes

```bash
stick list
```

### Show a Note

```bash
stick show 1
```

### Pin a Note

```bash
stick pin 1
```

### Search Notes

```bash
stick search milk
```

### Edit a Note

```bash
stick edit 1
```

### Delete a Note

```bash
stick delete 1
```

---

## Shell Startup Panel

Display your notes automatically whenever you open a terminal.

### Enable the Startup Panel

For Fish:

```bash
stick init fish
```

For Bash:

```bash
stick init bash
```

For Zsh:

```bash
stick init zsh
```

This adds a managed block to your shell startup file.

When you open a new terminal, Stick displays a preview of your latest note.

### Customize the Startup Panel

Control how many notes appear:

```bash
stick startup --limit 3
```

The limit can be set between 1 and 3 and is saved to the configuration file.

### Remove the Startup Hook

For Fish:

```bash
stick uninstall-shell fish
```

For Bash:

```bash
stick uninstall-shell bash
```

For Zsh:

```bash
stick uninstall-shell zsh
```

---

## Configuration

View all settings:

```bash
stick config
```

### Startup Panel Settings

Enable or disable the startup panel:

```bash
stick config startup enable
stick config startup disable
```

Show only pinned notes:

```bash
stick config startup pinned_only on
```

Show a hint when no notes exist:

```bash
stick config startup show_when_empty on
```

Set the startup note limit:

```bash
stick startup --limit 2
```

### Configuration File

```text
~/.config/stick/config.yaml
```

---

## Commands

| Command | Description |
|---------|-------------|
| `stick "text"` | Quickly create a note |
| `stick add [text]` | Create a new note |
| `stick add --title "T" "text"` | Create a note with a custom title |
| `stick list` | List all notes |
| `stick list --all` | Include archived notes |
| `stick list --pin` | Show only pinned notes |
| `stick show <id>` | Show a note with all topics |
| `stick edit <id>` | Edit a note in `$EDITOR` |
| `stick edit <id> --title "New"` | Rename a note |
| `stick edit <id> --topic "a,b"` | Replace all topics |
| `stick edit <id> --add "a,b"` | Append topics |
| `stick delete <id>` | Delete a note |
| `stick search <query>` | Search note titles and topics |
| `stick pin <id>` | Pin a note |
| `stick unpin <id>` | Unpin a note |
| `stick archive <id>` | Archive a note |
| `stick unarchive <id>` | Restore an archived note |
| `stick init <shell>` | Install the startup hook |
| `stick uninstall-shell <shell>` | Remove the startup hook |
| `stick startup` | Render the startup panel |
| `stick config` | Show configuration |
| `stick doctor` | Check database, config, and shell integration |
| `stick version` | Print version information |

---

## JSON Output

Stick supports machine-readable JSON output.

```bash
stick --json list
stick --json show 1
stick --json search docker
```

This is useful for shell scripts and automation.

---

## Data Storage

All data stays on your local machine.

| Data | Location |
|------|----------|
| Database | `~/.local/share/stick/notes.db` |
| Configuration | `~/.config/stick/config.yaml` |
| Backups | `~/.local/share/stick/backups/` |

The SQLite database uses a pure-Go driver, so CGO is not required.

Backups are created before shell file modifications.

---

## Development

### Clone the Repository

```bash
git clone https://github.com/hareshkhan01/sticky-notes.git
cd sticky-notes
```

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Go Vet

```bash
make vet
```

### Format Code

```bash
make fmt
```

### Install Locally

```bash
make install
```

The binary is installed to `$GOPATH/bin/stick`.

---

## Releases

Precompiled binaries are published through GitHub Releases.

**[View all releases](https://github.com/hareshkhan01/sticky-notes/releases)**

---

## License

This project is licensed under the MIT License.
