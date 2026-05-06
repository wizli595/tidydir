package executor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wizli595/tidydir/internal/action"
)

func TestRun_MoveAction(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "file.txt")
	os.WriteFile(src, []byte("hello"), 0644)

	dest := filepath.Join(root, "_docs", "file.txt")
	actions := []action.Action{
		{Type: action.ActionMove, Source: src, Dest: dest},
	}

	if err := Run(actions, root); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Error("dest file should exist after move")
	}
	if _, err := os.Stat(src); err == nil {
		t.Error("source file should not exist after move")
	}
}

func TestRun_DeleteAction(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "junk.tmp")
	os.WriteFile(src, []byte("trash"), 0644)

	actions := []action.Action{
		{Type: action.ActionDelete, Source: src},
	}

	if err := Run(actions, root); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(src); err == nil {
		t.Error("source should not exist after delete")
	}

	// Should be in trash
	trashEntries, _ := filepath.Glob(filepath.Join(root, ".tidydir_trash", "*", "junk.tmp"))
	if len(trashEntries) == 0 {
		t.Error("deleted file should be in .tidydir_trash")
	}
}

func TestRun_RenameAction(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "Old Name.txt")
	os.WriteFile(src, []byte("data"), 0644)

	dest := filepath.Join(root, "old-name.txt")
	actions := []action.Action{
		{Type: action.ActionRename, Source: src, Dest: dest},
	}

	if err := Run(actions, root); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Error("renamed file should exist")
	}
}

func TestRun_WritesUndoLog(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "file.txt")
	os.WriteFile(src, []byte("data"), 0644)

	dest := filepath.Join(root, "_docs", "file.txt")
	actions := []action.Action{
		{Type: action.ActionMove, Source: src, Dest: dest},
	}

	Run(actions, root)

	logPath := filepath.Join(root, ".tidydir_undo.json")
	if _, err := os.Stat(logPath); err != nil {
		t.Error("undo log should be created")
	}
}

func TestUndo_ReversesMove(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "file.txt")
	os.WriteFile(src, []byte("content"), 0644)

	dest := filepath.Join(root, "_docs", "file.txt")
	actions := []action.Action{
		{Type: action.ActionMove, Source: src, Dest: dest},
	}

	Run(actions, root)

	if err := Undo(root); err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Error("file should be restored to original location")
	}
	if _, err := os.Stat(dest); err == nil {
		t.Error("dest should not exist after undo")
	}
}

func TestUndo_RestoresDelete(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "junk.tmp")
	os.WriteFile(src, []byte("trash"), 0644)

	actions := []action.Action{
		{Type: action.ActionDelete, Source: src},
	}

	Run(actions, root)
	if err := Undo(root); err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Error("deleted file should be restored after undo")
	}
}

func TestUndo_CleansEmptyDirs(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "file.txt")
	os.WriteFile(src, []byte("data"), 0644)

	dest := filepath.Join(root, "sub", "deep", "file.txt")
	actions := []action.Action{
		{Type: action.ActionMove, Source: src, Dest: dest},
	}

	Run(actions, root)
	Undo(root)

	if _, err := os.Stat(filepath.Join(root, "sub")); err == nil {
		t.Error("empty sub/ directory should be cleaned up after undo")
	}
}

func TestUndo_RemovesLogAndTrash(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "junk.tmp")
	os.WriteFile(src, []byte("x"), 0644)

	Run([]action.Action{{Type: action.ActionDelete, Source: src}}, root)
	Undo(root)

	if _, err := os.Stat(filepath.Join(root, ".tidydir_undo.json")); err == nil {
		t.Error("undo log should be removed after undo")
	}
	if _, err := os.Stat(filepath.Join(root, ".tidydir_trash")); err == nil {
		t.Error("trash dir should be removed after undo")
	}
}

func TestUndo_NoLog(t *testing.T) {
	root := t.TempDir()
	err := Undo(root)
	if err == nil {
		t.Error("expected error when no undo log exists")
	}
}
