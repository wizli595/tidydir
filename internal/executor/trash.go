package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// MoveToSystemTrash moves a file or directory to the OS trash.
// Returns an error if the platform is unsupported or the operation fails.
func MoveToSystemTrash(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		return trashWindows(absPath)
	case "darwin":
		return trashDarwin(absPath)
	default:
		return trashLinux(absPath)
	}
}

func trashWindows(path string) error {
	// Use PowerShell with VisualBasic FileSystem to send to Recycle Bin
	escaped := strings.ReplaceAll(path, "'", "''")
	script := fmt.Sprintf(
		`Add-Type -AssemblyName Microsoft.VisualBasic; `+
			`[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`,
		escaped,
	)

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		script = fmt.Sprintf(
			`Add-Type -AssemblyName Microsoft.VisualBasic; `+
				`[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteDirectory('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`,
			escaped,
		)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("recycle bin: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func trashDarwin(path string) error {
	escaped := strings.ReplaceAll(path, `"`, `\"`)
	script := fmt.Sprintf(
		`tell application "Finder" to delete POSIX file "%s"`, escaped,
	)
	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("macos trash: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func trashLinux(path string) error {
	// Use XDG trash spec: move to ~/.local/share/Trash/
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home dir: %w", err)
	}

	trashFiles := filepath.Join(home, ".local", "share", "Trash", "files")
	trashInfo := filepath.Join(home, ".local", "share", "Trash", "info")

	if err := os.MkdirAll(trashFiles, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(trashInfo, 0755); err != nil {
		return err
	}

	base := filepath.Base(path)
	dest := filepath.Join(trashFiles, base)

	// Handle name conflicts
	if _, err := os.Stat(dest); err == nil {
		for i := 1; ; i++ {
			dest = filepath.Join(trashFiles, fmt.Sprintf("%s.%d", base, i))
			if _, err := os.Stat(dest); err != nil {
				break
			}
		}
	}

	if err := os.Rename(path, dest); err != nil {
		return fmt.Errorf("moving to trash: %w", err)
	}

	// Write .trashinfo file
	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		path, "now")
	infoPath := filepath.Join(trashInfo, filepath.Base(dest)+".trashinfo")
	os.WriteFile(infoPath, []byte(infoContent), 0644)

	return nil
}
