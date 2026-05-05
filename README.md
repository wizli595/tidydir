# tidydir

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20|%20macOS%20|%20Linux-blue)]()
[![GitHub release](https://img.shields.io/github/v/release/wizli595/tidydir?include_prereleases)](https://github.com/wizli595/tidydir/releases)

A CLI tool that scans messy directories, classifies files and projects by type, proposes an organization plan, and executes only what you approve.

## Features

- **Smart classification** — detects dev projects (Go, Node, Java, Flutter, .NET, Python, Rust, Docker, Django), documents, media, fonts, archives, duplicates, and junk files
- **Non-destructive** — deletes go to a trash folder, every action is logged for undo
- **Interactive confirmation** — approve all, none, or pick individually
- **Configurable** — YAML-based rules, custom move patterns, ignore lists, configurable folder names
- **Recursive scanning** — optional depth control with `--depth` flag
- **Dry-run mode** — preview changes without executing

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
tidydir scan ~/Documents --depth 2
```

Shows the proposed plan without making any changes.

### Organize (scan + confirm + execute)

```bash
tidydir organize ~/Documents
tidydir organize ~/Documents --dry-run
tidydir organize ~/Documents --depth 1
```

Scans, shows the plan, asks for confirmation, then executes approved actions.

Confirmation modes:
- `a` — approve all
- `n` — cancel
- `i` — interactive (approve/reject each action individually)

### Undo

```bash
tidydir undo ~/Documents
```

Reverses the last organize operation. Restores all moved files and recovers trashed items.

### Flags

| Flag | Command | Description |
|------|---------|-------------|
| `--depth N` | scan, organize | Recursion depth (0 = top-level only) |
| `--dry-run` | organize | Show plan without executing |
| `--help` | all | Show help for any command |

## How it works

```
Scanner -> Classifiers (Strategy pattern) -> Planner -> UI (confirm) -> Executor
```

1. **Scanner** reads directory entries, respects ignore patterns, supports recursive depth
2. **Classifiers** run in priority order (junk > duplicates > projects > file types). First match wins.
3. **Planner** converts classifications into concrete actions (move, delete, rename), applies custom rules
4. **UI** displays the plan with colored output and collects user approval
5. **Executor** runs approved actions, writes undo log, cleans up on undo

## Classification rules

| Category | Detection method |
|----------|-----------------|
| Project | Marker files: `go.mod`, `package.json`, `Cargo.toml`, `pom.xml`, `pubspec.yaml`, `*.csproj`, `docker-compose.yml`, `manage.py`, etc. |
| Document | Extensions: `.pdf`, `.docx`, `.xlsx`, `.csv`, `.html`, `.txt`, `.pptx` |
| Media | Extensions: `.png`, `.jpg`, `.mp4`, `.mp3`, `.svg`, `.gif`, `.webp` |
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
    docker/
  _docs/
  _media/
  _fonts/
  _archives/
  .tidydir_trash/    # soft-deleted files (duplicates, junk)
  .tidydir_undo.json # undo log
```

## Configuration

Place a `tidydir.yaml` or `.tidydir.yaml` in the target directory to override defaults:

```yaml
folders:
  project: "projects"
  document: "_docs"
  media: "_media"
  font: "_fonts"
  archive: "_archives"

project_markers:
  - file: "docker-compose.yml"
    type: "docker"
  - file: "manage.py"
    type: "django"

ignore:
  - "desktop.ini"
  - "*.lnk"
  - "My Music"

custom_rules:
  - pattern: "*.sketch"
    dest: "_design"
```

## Architecture

```
tidydir/
  cmd/              # CLI commands (cobra)
  internal/
    scanner/        # Directory walking with ignore and depth
    classifier/     # Strategy pattern: each rule implements the Classifier interface
    config/         # YAML config loader with defaults
    planner/        # Converts classifications to actions, applies custom rules
    action/         # Action type definitions
    executor/       # Runs actions, writes undo log, handles rollback
    ui/             # Terminal output and confirmation prompts
  config/           # Default YAML rule definitions
```

## Dependencies

- [cobra](https://github.com/spf13/cobra) — CLI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [yaml.v3](https://gopkg.in/yaml.v3) — Configuration parsing

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
