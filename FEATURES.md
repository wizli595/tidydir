# Features and Phases

## Phase 1 — Core (MVP)

- [x] Directory scanning (top-level entries)
- [x] Project detection by marker files (go.mod, package.json, Cargo.toml, etc.)
- [x] File type classification by extension
- [x] Duplicate detection (zip+folder pairs, copy patterns)
- [x] Junk file detection
- [x] Action planning (move, delete)
- [x] Colored terminal output
- [x] Interactive confirmation (all / none / per-item)
- [x] Undo log with full reversal
- [x] Trash-based deletion (non-destructive)

## Phase 2 — Configuration and Flexibility

- [x] Load rules from YAML config at runtime
- [x] Custom move rules (pattern -> destination)
- [x] Ignore patterns (skip specific files/folders)
- [x] Configurable target folder names
- [x] Recursive scanning (optional --depth flag)
- [x] Dry-run flag (`--dry-run` on organize)
- [x] Extra project markers from config (docker-compose, Makefile, manage.py)
- [x] Undo accepts path argument
- [x] Undo cleans up empty directories and trash

## Phase 3 — Intelligence

- [x] File size awareness (flag large files, configurable threshold)
- [x] Last-modified date grouping (flag old projects, configurable days)
- [x] Git repo detection (warn about uncommitted changes before moving)
- [x] Naming convention suggestions (kebab-case normalization)
- [x] Batch rename support (rename actions in plan, approve individually)
- [x] Detect orphaned node_modules / vendor / build dirs and report size

## Phase 4 — UX Polish

- [x] Bubbletea interactive TUI (navigate plan with arrow keys, space to toggle)
- [x] Summary report after execution (moved X, deleted Y, freed Z MB)
- [x] Color themes (auto dark/light terminal detection via termenv)
- [x] Shell completions (bash, zsh, fish, powershell)
- [x] `tidydir watch` — monitor a directory with fsnotify and report when messy

## Phase 5 — Advanced

- [x] Plugin system (custom classifiers via YAML config with extension/pattern matching)
- [x] Profile support (--profile flag, different rule sets per use case)
- [x] Integration with system trash (--system-trash flag: Windows Recycle Bin, macOS Trash, Linux XDG)
- [x] Export plan to JSON/CSV (--format json|csv on scan command)
- [x] Multi-directory support (`tidydir scan ~/Documents ~/Downloads`)
- [x] Scheduled auto-scan with desktop notifications (`tidydir schedule`)
