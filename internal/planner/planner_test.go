package planner

import (
	"path/filepath"
	"testing"

	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/scanner"
)

var defaultFolders = map[string]string{
	"project":  "projects",
	"document": "_docs",
	"media":    "_media",
	"font":     "_fonts",
	"archive":  "_archives",
}

func TestPlan_MoveProject(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "my-app", Path: filepath.Join(root, "my-app"), IsDir: true},
			Category: classifier.CatProject,
			SubType:  "go",
			Reason:   "found go.mod",
		},
	}

	actions := Plan(classifications, root, defaultFolders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Type != action.ActionMove {
		t.Errorf("got type %s, want move", actions[0].Type)
	}
	if actions[0].Dest != filepath.Join(root, "projects", "go", "my-app") {
		t.Errorf("got dest %s, want projects/go/my-app", actions[0].Dest)
	}
}

func TestPlan_MoveDocument(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "report.pdf", Path: filepath.Join(root, "report.pdf"), Ext: ".pdf"},
			Category: classifier.CatDocument,
			Reason:   "extension .pdf",
		},
	}

	actions := Plan(classifications, root, defaultFolders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Dest != filepath.Join(root, "_docs", "report.pdf") {
		t.Errorf("got dest %s, want _docs/report.pdf", actions[0].Dest)
	}
}

func TestPlan_DeleteJunk(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: ".DS_Store", Path: filepath.Join(root, ".DS_Store")},
			Category: classifier.CatJunk,
			Reason:   "system junk file",
		},
	}

	actions := Plan(classifications, root, defaultFolders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Type != action.ActionDelete {
		t.Errorf("got type %s, want delete", actions[0].Type)
	}
}

func TestPlan_DeleteDuplicate(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "file (1).txt", Path: filepath.Join(root, "file (1).txt")},
			Category: classifier.CatDuplicate,
			Reason:   "looks like a copy",
		},
	}

	actions := Plan(classifications, root, defaultFolders, nil)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Type != action.ActionDelete {
		t.Errorf("got type %s, want delete", actions[0].Type)
	}
}

func TestPlan_CustomRulesFirst(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "design.sketch", Path: filepath.Join(root, "design.sketch"), Ext: ".sketch"},
			Category: classifier.CatUnknown,
			Reason:   "no rule matched",
		},
	}
	rules := []config.CustomRule{
		{Pattern: "*.sketch", Dest: "_design"},
	}

	actions := Plan(classifications, root, defaultFolders, rules)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Dest != filepath.Join(root, "_design", "design.sketch") {
		t.Errorf("got dest %s, want _design/design.sketch", actions[0].Dest)
	}
}

func TestPlan_SkipsAlreadyInPlace(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "report.pdf", Path: filepath.Join(root, "_docs", "report.pdf"), Ext: ".pdf"},
			Category: classifier.CatDocument,
			Reason:   "extension .pdf",
		},
	}

	actions := Plan(classifications, root, defaultFolders, nil)
	if len(actions) != 0 {
		t.Errorf("expected 0 actions for file already in place, got %d", len(actions))
	}
}

func TestPlan_UnknownCategory(t *testing.T) {
	root := "/tmp/test"
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "random.xyz", Path: filepath.Join(root, "random.xyz")},
			Category: classifier.CatUnknown,
			Reason:   "no rule matched",
		},
	}

	actions := Plan(classifications, root, defaultFolders, nil)
	if len(actions) != 0 {
		t.Errorf("expected 0 actions for unknown, got %d", len(actions))
	}
}

func TestPlanRenames_NormalizesName(t *testing.T) {
	root := t.TempDir()
	entries := []scanner.Entry{
		{Name: "My Document.pdf", Path: filepath.Join(root, "My Document.pdf"), Ext: ".pdf"},
	}

	renames := PlanRenames(entries, root, nil)
	if len(renames) != 1 {
		t.Fatalf("expected 1 rename, got %d", len(renames))
	}
	if renames[0].Type != action.ActionRename {
		t.Errorf("got type %s, want rename", renames[0].Type)
	}
	if renames[0].Dest != filepath.Join(root, "my-document.pdf") {
		t.Errorf("got dest %s, want my-document.pdf", renames[0].Dest)
	}
}

func TestPlanRenames_SkipsHandledEntries(t *testing.T) {
	root := t.TempDir()
	entries := []scanner.Entry{
		{Name: "My File.txt", Path: filepath.Join(root, "My File.txt"), Ext: ".txt"},
	}
	existing := []action.Action{
		{Type: action.ActionMove, Source: filepath.Join(root, "My File.txt"), Dest: filepath.Join(root, "other")},
	}

	renames := PlanRenames(entries, root, existing)
	if len(renames) != 0 {
		t.Errorf("expected 0 renames for handled entry, got %d", len(renames))
	}
}

func TestPlanRenames_SkipsCleanNames(t *testing.T) {
	root := t.TempDir()
	entries := []scanner.Entry{
		{Name: "clean-name.txt", Path: filepath.Join(root, "clean-name.txt"), Ext: ".txt"},
	}

	renames := PlanRenames(entries, root, nil)
	if len(renames) != 0 {
		t.Errorf("expected 0 renames for clean name, got %d", len(renames))
	}
}

func TestPlanRenames_SkipsDirs(t *testing.T) {
	root := t.TempDir()
	entries := []scanner.Entry{
		{Name: "My Folder", Path: filepath.Join(root, "My Folder"), IsDir: true},
	}

	renames := PlanRenames(entries, root, nil)
	if len(renames) != 0 {
		t.Errorf("expected 0 renames for directory, got %d", len(renames))
	}
}

func TestPlanRenames_SkipsNestedEntries(t *testing.T) {
	root := t.TempDir()
	entries := []scanner.Entry{
		{Name: "Bad Name.txt", Path: filepath.Join(root, "sub", "Bad Name.txt"), Ext: ".txt"},
	}

	renames := PlanRenames(entries, root, nil)
	if len(renames) != 0 {
		t.Errorf("expected 0 renames for nested entry, got %d", len(renames))
	}
}
