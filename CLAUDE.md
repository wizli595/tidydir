# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, Run

```bash
go build -o tidydir.exe .     # build
go test ./...                  # run all tests (58 tests across 8 packages)
go mod tidy                    # fix/update dependencies
./tidydir.exe scan <path>              # preview plan + insights
./tidydir.exe scan <path> --format json  # export plan as JSON
./tidydir.exe scan <p1> <p2>           # scan multiple directories
./tidydir.exe organize <path>          # execute with TUI confirmation
./tidydir.exe organize <path> --system-trash  # use OS trash
./tidydir.exe organize <path> --profile work  # use named profile
./tidydir.exe undo <path>              # revert last run
./tidydir.exe watch <path>             # monitor directory for changes
./tidydir.exe schedule <path>          # setup scheduled auto-scan
./tidydir.exe completion bash          # generate shell completions
```

Run a single package's tests: `go test ./internal/scanner/`

## Architecture

The tool follows a **pipeline pattern** with a **strategy pattern** for classification:

```
Scanner → Classifiers → Insights → Planner → UI (TUI confirm) → Executor → Summary
```

- **`cmd/`** — Cobra CLI commands (`scan`, `organize`, `undo`, `watch`, `schedule`, `completion`). Each wires the pipeline together.
- **`internal/scanner/`** — Walks a directory with configurable depth and ignore patterns. Returns flat `[]Entry` with Name, Path, IsDir, Size, Ext, ModTime.
- **`internal/classifier/`** — Strategy pattern. Five classifiers run in priority order (junk > duplicate > project > custom > filetype). First match wins. All implement `Classifier` interface with `Classify(entry, allEntries) *Classification`. Custom classifiers are loaded from config YAML.
- **`internal/insights/`** — Analyzes entries and classifications to produce warnings: large files, old projects, dirty git repos, heavy dependency dirs, naming issues. Also exports `NormalizeName()` for kebab-case conversion.
- **`internal/config/`** — Loads YAML config with defaults. Searches `tidydir.yaml` in target dir, then `config/rules.yaml` next to binary. Supports profiles and custom classifiers. `LoadWithProfile()` applies named profile overrides.
- **`internal/planner/`** — Converts `[]Classification` into `[]Action` (move/delete/rename). `Plan()` applies custom rules first, then category-based rules. `PlanRenames()` generates rename actions for naming normalization.
- **`internal/action/`** — Action type definitions only (move, delete, rename).
- **`internal/executor/`** — Executes actions, returns `Stats` (moved/deleted/renamed/freed), writes JSON undo log. Deletes are non-destructive (moved to `.tidydir_trash/` or system trash via `--system-trash`). Undo reads log in reverse (LIFO).
- **`internal/export/`** — Exports action plans to JSON or CSV format for external review.
- **`internal/ui/`** — Bubbletea TUI for interactive approval (arrow keys, space toggle, checkboxes). ShowPlan, ShowInsights, ShowSummary. Auto dark/light theme via termenv.

## Key Design Rules

- Classifiers run in priority order — adding a new one means inserting it at the right position in `DefaultClassifiers()` in `classifier.go`.
- All deletes go to `.tidydir_trash/{date}/` by default, or system trash with `--system-trash`. Undo restores from local trash only.
- The planner receives config folders map and custom rules — it does not read config directly.
- Scanner returns a flat list, not a tree. Depth controls recursion but output is always flat.
- Config uses `yaml.Unmarshal` directly onto a struct with defaults already set — no merge logic.
- `executor.Run()` returns `(Stats, error)` — accepts optional `RunOptions` for system trash.
- `DefaultClassifiers()` is variadic — pass custom classifier rules as second arg.
- Multi-dir commands loop independently per path. Each gets its own config/profile.

## Dependencies

- **cobra** — CLI commands and flags
- **bubbletea** — Interactive TUI (action approval)
- **lipgloss** — Terminal styling
- **termenv** — Dark/light terminal detection
- **fsnotify** — File system watching (watch command)
- **yaml.v3** — Config parsing

## CI/CD

`.github/workflows/release.yml` builds 6 binaries (windows/darwin/linux x amd64/arm64) on every push to master. Version auto-generated from commit count.
