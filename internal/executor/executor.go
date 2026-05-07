package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wizli595/tidydir/internal/action"
)

// LogEntry records an executed action for undo.
type LogEntry struct {
	Type     action.ActionType `json:"type"`
	Source   string            `json:"source"`
	Dest     string            `json:"dest,omitempty"`
	BackupAt string            `json:"backup_at,omitempty"` // for deletes
}

const logFile = ".tidydir_undo.json"

// Stats holds execution counts for the summary report.
type Stats struct {
	Moved   int
	Deleted int
	Renamed int
	Freed   int64
}

// RunOptions configures execution behavior.
type RunOptions struct {
	SystemTrash bool // use OS trash instead of .tidydir_trash
}

// Run executes all approved actions and writes an undo log.
func Run(actions []action.Action, rootPath string, opts ...RunOptions) (Stats, error) {
	var log []LogEntry
	var stats Stats
	var opt RunOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	for _, a := range actions {
		switch a.Type {
		case action.ActionMove:
			destDir := filepath.Dir(a.Dest)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return stats, fmt.Errorf("creating dir %s: %w", destDir, err)
			}
			if err := os.Rename(a.Source, a.Dest); err != nil {
				return stats, fmt.Errorf("moving %s → %s: %w", a.Source, a.Dest, err)
			}
			log = append(log, LogEntry{Type: a.Type, Source: a.Source, Dest: a.Dest})
			stats.Moved++

		case action.ActionDelete:
			// Calculate freed space before trashing
			stats.Freed += CalcSize(a.Source)

			if opt.SystemTrash {
				if err := MoveToSystemTrash(a.Source); err != nil {
					return stats, fmt.Errorf("system trash %s: %w", a.Source, err)
				}
				log = append(log, LogEntry{Type: a.Type, Source: a.Source})
			} else {
				trashDir := filepath.Join(rootPath, ".tidydir_trash", time.Now().Format("2006-01-02"))
				if err := os.MkdirAll(trashDir, 0755); err != nil {
					return stats, fmt.Errorf("creating trash dir: %w", err)
				}
				backupPath := filepath.Join(trashDir, filepath.Base(a.Source))
				if err := os.Rename(a.Source, backupPath); err != nil {
					return stats, fmt.Errorf("trashing %s: %w", a.Source, err)
				}
				log = append(log, LogEntry{Type: a.Type, Source: a.Source, BackupAt: backupPath})
			}
			stats.Deleted++

		case action.ActionRename:
			if err := os.Rename(a.Source, a.Dest); err != nil {
				return stats, fmt.Errorf("renaming %s → %s: %w", a.Source, a.Dest, err)
			}
			log = append(log, LogEntry{Type: a.Type, Source: a.Source, Dest: a.Dest})
			stats.Renamed++
		}
	}

	logPath := filepath.Join(rootPath, logFile)
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return stats, fmt.Errorf("serializing undo log: %w", err)
	}
	return stats, os.WriteFile(logPath, data, 0644)
}

// CalcSize returns the total size in bytes for a file or directory.
// Uses WalkDir instead of Walk to avoid redundant os.Stat calls.
func CalcSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}
	var size int64
	filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if fi, err := d.Info(); err == nil {
			size += fi.Size()
		}
		return nil
	})
	return size
}

// Undo reverses the last organize operation using the undo log.
func Undo(rootPath string) error {
	logPath := filepath.Join(rootPath, logFile)
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("no undo log found: %w", err)
	}

	var log []LogEntry
	if err := json.Unmarshal(data, &log); err != nil {
		return fmt.Errorf("parsing undo log: %w", err)
	}

	// Reverse order
	for i := len(log) - 1; i >= 0; i-- {
		entry := log[i]
		switch entry.Type {
		case action.ActionMove, action.ActionRename:
			destDir := filepath.Dir(entry.Source)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("creating dir for undo: %w", err)
			}
			if err := os.Rename(entry.Dest, entry.Source); err != nil {
				return fmt.Errorf("undoing move %s: %w", entry.Dest, err)
			}
		case action.ActionDelete:
			if entry.BackupAt != "" {
				destDir := filepath.Dir(entry.Source)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					return fmt.Errorf("creating dir for restore: %w", err)
				}
				if err := os.Rename(entry.BackupAt, entry.Source); err != nil {
					return fmt.Errorf("restoring %s: %w", entry.Source, err)
				}
			}
		}
	}

	// Clean up empty directories created by organize
	cleanEmptyDirs(rootPath)

	// Remove the log file and trash dir after successful undo
	os.RemoveAll(filepath.Join(rootPath, ".tidydir_trash"))
	os.Remove(logPath)
	return nil
}

func cleanEmptyDirs(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(root, e.Name())
		cleanEmptyDirs(path)
		// Remove if now empty
		sub, _ := os.ReadDir(path)
		if len(sub) == 0 {
			os.Remove(path)
		}
	}
}
