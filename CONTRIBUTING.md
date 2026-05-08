# Contributing to tidydir

Thanks for your interest in contributing. This document explains how to get started.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/<your-user>/tidydir.git`
3. Create a branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit and push
7. Open a pull request against `master`

## Development Setup

```bash
# Requirements: Go 1.24+
go build -o tidydir.exe .
go test ./...
```

## Project Structure

```
cmd/              CLI commands (Cobra)
internal/
  scanner/        Directory walking
  classifier/     Strategy pattern classifiers
  config/         YAML config with profiles
  planner/        Classification -> Actions
  insights/       Intelligence analysis
  export/         JSON/CSV export
  action/         Action types
  executor/       Execute actions, undo, system trash
  ui/             Bubbletea TUI
```

## Adding a New Classifier

1. Create `internal/classifier/yourname.go` implementing the `Classifier` interface
2. Add it to `DefaultClassifiers()` in `classifier.go` at the right priority position
3. Add tests in `internal/classifier/yourname_test.go`
4. Run `go test ./internal/classifier/`

## Adding a New Command

1. Create `cmd/yourcommand.go`
2. Define a `cobra.Command` and register it in `init()` with `rootCmd.AddCommand()`
3. Update the README Commands section

## Code Guidelines

- Keep functions focused on a single responsibility
- Check errors. Do not ignore them unless explicitly documented why.
- Use `filepath.WalkDir` (not `filepath.Walk`) for directory traversal
- Do not add external dependencies without discussion
- All deletes must be non-destructive (trash-based)
- Run `go test ./...` before submitting

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if you changed behavior or added features
3. Keep commits focused. One logical change per commit.
4. Write clear commit messages describing what and why

## Reporting Bugs

Use the GitHub issue templates. Include:
- Operating system and architecture
- Go version
- Steps to reproduce
- Expected vs actual behavior

## Feature Requests

Open an issue with the "enhancement" label. Describe the use case, not just the solution.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
