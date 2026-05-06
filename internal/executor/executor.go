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

// Run executes all approved actions and writes an undo log.
func Run(actions []action.Action, rootPath string) (Stats, error) {
	var log []LogEntry
	var stats Stats

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
			if info, err := os.Stat(a.Source); err == nil {
				if info.IsDir() {
					stats.Freed += calcDirSize(a.Source)
				} else {
					stats.Freed += info.Size()
				}
			}
			trashDir := filepath.Join(rootPath, ".tidydir_trash", time.Now().Format("2006-01-02"))
			if err := os.MkdirAll(trashDir, 0755); err != nil {
				return stats, fmt.Errorf("creating trash dir: %w", err)
			}
			backupPath := filepath.Join(trashDir, filepath.Base(a.Source))
			if err := os.Rename(a.Source, backupPath); err != nil {
				return stats, fmt.Errorf("trashing %s: %w", a.Source, err)
			}
			log = append(log, LogEntry{Type: a.Type, Source: a.Source, BackupAt: backupPath})
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
	data, _ := json.MarshalIndent(log, "", "  ")
	return stats, os.WriteFile(logPath, data, 0644)
}

func calcDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
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
			// Move back: dest → source
			destDir := filepath.Dir(entry.Source)
			os.MkdirAll(destDir, 0755)
			if err := os.Rename(entry.Dest, entry.Source); err != nil {
				return fmt.Errorf("undoing move %s: %w", entry.Dest, err)
			}
		case action.ActionDelete:
			// Restore from trash
			if entry.BackupAt != "" {
				destDir := filepath.Dir(entry.Source)
				os.MkdirAll(destDir, 0755)
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
