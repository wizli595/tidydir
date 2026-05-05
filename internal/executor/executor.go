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

// Run executes all approved actions and writes an undo log.
func Run(actions []action.Action, rootPath string) error {
	var log []LogEntry

	for _, a := range actions {
		switch a.Type {
		case action.ActionMove:
			// Ensure destination directory exists
			destDir := filepath.Dir(a.Dest)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("creating dir %s: %w", destDir, err)
			}
			if err := os.Rename(a.Source, a.Dest); err != nil {
				return fmt.Errorf("moving %s → %s: %w", a.Source, a.Dest, err)
			}
			log = append(log, LogEntry{Type: a.Type, Source: a.Source, Dest: a.Dest})

		case action.ActionDelete:
			// Move to trash (backup dir) instead of permanent delete
			trashDir := filepath.Join(rootPath, ".tidydir_trash", time.Now().Format("2006-01-02"))
			if err := os.MkdirAll(trashDir, 0755); err != nil {
				return fmt.Errorf("creating trash dir: %w", err)
			}
			backupPath := filepath.Join(trashDir, filepath.Base(a.Source))
			if err := os.Rename(a.Source, backupPath); err != nil {
				return fmt.Errorf("trashing %s: %w", a.Source, err)
			}
			log = append(log, LogEntry{Type: a.Type, Source: a.Source, BackupAt: backupPath})

		case action.ActionRename:
			if err := os.Rename(a.Source, a.Dest); err != nil {
				return fmt.Errorf("renaming %s → %s: %w", a.Source, a.Dest, err)
			}
			log = append(log, LogEntry{Type: a.Type, Source: a.Source, Dest: a.Dest})
		}
	}

	// Write undo log
	logPath := filepath.Join(rootPath, logFile)
	data, _ := json.MarshalIndent(log, "", "  ")
	return os.WriteFile(logPath, data, 0644)
}

// Undo reverses the last organize operation using the undo log.
func Undo() error {
	// Find the log file in current directory or home
	logPath := logFile
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

	// Remove the log file after successful undo
	os.Remove(logPath)
	return nil
}
