# Stick

Sticky notes for your terminal. Each note is a page holding up to 5 topics with a 200-character limit per topic. Notes live in a local SQLite database — nothing leaves your machine.

## Install

```bash
git clone https://github.com/hareshkhan01/sticky-notes.git
cd sticky-notes
make install
```

Requires Go 1.21+. The binary is installed to `$GOPATH/bin/stick`.

## Quick Start

```bash
stick "remember the milk"                        # quick capture
stick add --title "Errands" "buy milk
post letter"                                     # multi-topic note
stick pin 1                                      # pin to top
stick list                                       # see all notes
stick show 1                                     # view a note
stick search milk                                # search titles and topics
stick edit 1                                     # open in $EDITOR
stick delete 1                                   # remove a note
```

## Shell Startup Panel

Show your notes every time you open a terminal:

```bash
stick init fish        # for fish shell
stick init bash        # for bash
stick init zsh         # for zsh
```

This appends a managed block to your shell startup file. It shows a preview of your latest note on each new terminal. To change how many notes appear:


<img width="461" height="300" alt="image" src="https://github.com/user-attachments/assets/25f4d7b7-b5c2-4f24-93a6-b84dfd5eda81" />



```bash
stick startup --limit 3    # show up to 3 notes (max 3, saved to config)
```

To remove the startup terminal hook:

```bash
stick uninstall-shell fish
```

## Configuration

```bash
stick config                                    # show all settings
stick config startup enable                     # enable startup panel
stick config startup disable                    # disable startup panel
stick config startup pinned_only on             # show only pinned notes at startup
stick config startup show_when_empty on         # show hint when no notes exist
stick startup --limit 2                         # set how many notes to show (1-3)
```

Config file location: `~/.config/stick/config.yaml`

## Commands

| Command | Description |
|---------|-------------|
| `stick "text"` | Quick capture — creates a note from arguments |
| `stick add [text]` | Create a new note |
| `stick add --title "T" "text"` | Create a note with a custom title |
| `stick list` | List all notes |
| `stick list --all` | Include archived notes |
| `stick list --pin` | Show only pinned notes |
| `stick show <id>` | Show a note with all topics |
| `stick edit <id>` | Edit a note in $EDITOR |
| `stick edit <id> --title "New"` | Rename a note |
| `stick edit <id> --topic "a,b"` | Replace all topics |
| `stick edit <id> --add "a,b"` | Append topics |
| `stick delete <id>` | Delete a note |
| `stick search <query>` | Search note titles and topics |
| `stick pin <id>` | Pin a note to the top |
| `stick unpin <id>` | Unpin a note |
| `stick archive <id>` | Archive a note |
| `stick unarchive <id>` | Restore an archived note |
| `stick init <shell>` | Install the startup hook |
| `stick uninstall-shell <shell>` | Remove the startup hook |
| `stick startup` | Render the startup panel |
| `stick config` | Show configuration |
| `stick doctor` | Check database, config, and shell integration |
| `stick version` | Print version information |

## JSON Output

Any command can output machine-readable JSON:

```bash
stick --json list
stick --json show 1
stick --json search docker
```

## Data

- **Database**: `~/.local/share/stick/notes.db` (SQLite, pure Go — no CGO required)
- **Config**: `~/.config/stick/config.yaml`
- **Backups**: `~/.local/share/stick/backups/` (created before shell file modifications)

## Development

```bash
make build       # build the binary
make test        # run all tests
make vet         # run go vet
make fmt         # format code
make install     # build and install to $GOPATH/bin
```
