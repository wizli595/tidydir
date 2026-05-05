# tidydir

A CLI tool that scans messy directories, classifies files and projects by type, proposes an organization plan, and executes only what you approve.

## Features

- **Smart classification** — detects dev projects (Go, Node, Java, Flutter, .NET, Python, Rust), documents, media, fonts, archives, duplicates, and junk files
- **Non-destructive** — deletes go to a trash folder, every action is logged for undo
- **Interactive confirmation** — approve all, none, or pick individually
- **Extensible rules** — add custom classifiers via the strategy pattern, or configure via YAML

## Install

```bash
go install github.com/wizli595/tidydir@latest
```

Or build from source:

```bash
git clone https://github.com/wizli595/tidydir.git
cd tidydir
go build -o tidydir .
```

## Usage

### Scan (preview only)

```bash
tidydir scan ~/Documents
```

Shows the proposed plan without making any changes.

### Organize (scan + confirm + execute)

```bash
tidydir organize ~/Documents
```

Scans, shows the plan, asks for confirmation, then executes approved actions.

Confirmation modes:
- `a` — approve all
- `n` — cancel
- `i` — interactive (approve/reject each action individually)

### Undo

```bash
tidydir undo
```

Reverses the last organize operation using the saved undo log.

## How it works

```
Scanner -> Classifiers (Strategy pattern) -> Planner -> UI (confirm) -> Executor
```

1. **Scanner** reads the top-level directory entries
2. **Classifiers** run in priority order (junk > duplicates > projects > file types). First match wins.
3. **Planner** converts classifications into concrete actions (move, delete, rename)
4. **UI** displays the plan with colored output and collects user approval
5. **Executor** runs approved actions and writes an undo log

## Classification rules

| Category | Detection method |
|----------|-----------------|
| Project | Marker files: `go.mod`, `package.json`, `Cargo.toml`, `pom.xml`, `pubspec.yaml`, `*.csproj`, etc. |
| Document | Extensions: `.pdf`, `.docx`, `.xlsx`, `.csv`, `.html`, `.txt` |
| Media | Extensions: `.png`, `.jpg`, `.mp4`, `.mp3`, `.svg` |
| Font | Extensions: `.ttf`, `.otf`, `.woff`, `.woff2` |
| Archive | Extensions: `.zip`, `.tar`, `.gz`, `.rar`, `.7z` |
| Duplicate | Zip with matching extracted folder, `(1)` / `- Copy` in filename |
| Junk | `.DS_Store`, `Thumbs.db`, `node_modules`, `.cache` |

## Output structure

After organizing, your directory will look like:

```
Documents/
  projects/
    go/
    node/
    java/
    flutter/
  _docs/
  _media/
  _fonts/
  _archives/
  .tidydir_trash/    # soft-deleted files (duplicates, junk)
  .tidydir_undo.json # undo log
```

## Configuration

Edit `config/rules.yaml` to customize target folders, add project markers, define ignore patterns, or add custom move rules.

## Architecture

```
tidydir/
  cmd/           # CLI commands (cobra)
  internal/
    scanner/     # Directory walking
    classifier/  # Strategy pattern: each rule implements the Classifier interface
    planner/     # Converts classifications to actions
    action/      # Action type definitions
    executor/    # Runs actions, writes undo log
    ui/          # Terminal output and confirmation prompts
  config/        # YAML rule definitions
```

## Dependencies

- [cobra](https://github.com/spf13/cobra) — CLI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling

## License

MIT
