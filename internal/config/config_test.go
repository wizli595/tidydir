package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	cfg := Load(dir)

	if cfg.Folders["project"] != "projects" {
		t.Errorf("project folder = %q, want projects", cfg.Folders["project"])
	}
	if cfg.Folders["document"] != "_docs" {
		t.Errorf("document folder = %q, want _docs", cfg.Folders["document"])
	}
	if cfg.LargeFileMB != 100 {
		t.Errorf("large_file_mb = %d, want 100", cfg.LargeFileMB)
	}
	if cfg.OldProjectDays != 180 {
		t.Errorf("old_project_days = %d, want 180", cfg.OldProjectDays)
	}
	if len(cfg.Ignore) == 0 {
		t.Error("expected default ignore patterns")
	}
}

func TestLoad_YamlOverride(t *testing.T) {
	dir := t.TempDir()
	yaml := `
folders:
  project: "dev"
  document: "docs"
ignore:
  - "*.tmp"
large_file_mb: 50
old_project_days: 90
custom_rules:
  - pattern: "*.sketch"
    dest: "_design"
`
	os.WriteFile(filepath.Join(dir, "tidydir.yaml"), []byte(yaml), 0644)
	cfg := Load(dir)

	if cfg.Folders["project"] != "dev" {
		t.Errorf("project folder = %q, want dev", cfg.Folders["project"])
	}
	if cfg.Folders["document"] != "docs" {
		t.Errorf("document folder = %q, want docs", cfg.Folders["document"])
	}
	if cfg.LargeFileMB != 50 {
		t.Errorf("large_file_mb = %d, want 50", cfg.LargeFileMB)
	}
	if cfg.OldProjectDays != 90 {
		t.Errorf("old_project_days = %d, want 90", cfg.OldProjectDays)
	}
	if len(cfg.Ignore) != 1 || cfg.Ignore[0] != "*.tmp" {
		t.Errorf("ignore = %v, want [*.tmp]", cfg.Ignore)
	}
	if len(cfg.CustomRules) != 1 || cfg.CustomRules[0].Pattern != "*.sketch" {
		t.Errorf("custom rules not loaded correctly: %v", cfg.CustomRules)
	}
}

func TestLoad_DotFile(t *testing.T) {
	dir := t.TempDir()
	yaml := `
folders:
  media: "images"
`
	os.WriteFile(filepath.Join(dir, ".tidydir.yaml"), []byte(yaml), 0644)
	cfg := Load(dir)

	if cfg.Folders["media"] != "images" {
		t.Errorf("media folder = %q, want images", cfg.Folders["media"])
	}
}

func TestLoad_ProjectMarkers(t *testing.T) {
	dir := t.TempDir()
	yaml := `
project_markers:
  - file: "docker-compose.yml"
    type: "docker"
  - file: "manage.py"
    type: "django"
`
	os.WriteFile(filepath.Join(dir, "tidydir.yaml"), []byte(yaml), 0644)
	cfg := Load(dir)

	if len(cfg.ProjectMarkers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(cfg.ProjectMarkers))
	}
	if cfg.ProjectMarkers[0].Type != "docker" {
		t.Errorf("first marker type = %q, want docker", cfg.ProjectMarkers[0].Type)
	}
}

func TestLoadWithProfile(t *testing.T) {
	dir := t.TempDir()
	yaml := `
folders:
  project: "projects"
profiles:
  work:
    folders:
      project: "work-projects"
    large_file_mb: 500
  personal:
    folders:
      project: "hobby"
    ignore:
      - "*.bak"
`
	os.WriteFile(filepath.Join(dir, "tidydir.yaml"), []byte(yaml), 0644)

	cfg := LoadWithProfile(dir, "work")
	if cfg.Folders["project"] != "work-projects" {
		t.Errorf("expected work-projects, got %q", cfg.Folders["project"])
	}
	if cfg.LargeFileMB != 500 {
		t.Errorf("expected 500, got %d", cfg.LargeFileMB)
	}

	cfg2 := LoadWithProfile(dir, "personal")
	if cfg2.Folders["project"] != "hobby" {
		t.Errorf("expected hobby, got %q", cfg2.Folders["project"])
	}
	if len(cfg2.Ignore) != 1 || cfg2.Ignore[0] != "*.bak" {
		t.Errorf("expected [*.bak], got %v", cfg2.Ignore)
	}
}

func TestLoadWithProfile_Unknown(t *testing.T) {
	dir := t.TempDir()
	cfg := LoadWithProfile(dir, "nonexistent")
	if cfg.Folders["project"] != "projects" {
		t.Errorf("expected default, got %q", cfg.Folders["project"])
	}
}

func TestLoad_CustomClassifiers(t *testing.T) {
	dir := t.TempDir()
	yaml := `
custom_classifiers:
  - name: "data-files"
    extensions: [".parquet", ".avro"]
    category: "document"
    subtype: "data"
  - name: "logs"
    patterns: ["*.log"]
    category: "junk"
`
	os.WriteFile(filepath.Join(dir, "tidydir.yaml"), []byte(yaml), 0644)
	cfg := Load(dir)

	if len(cfg.CustomClassifiers) != 2 {
		t.Fatalf("expected 2 custom classifiers, got %d", len(cfg.CustomClassifiers))
	}
	if cfg.CustomClassifiers[0].Name != "data-files" {
		t.Errorf("expected 'data-files', got %q", cfg.CustomClassifiers[0].Name)
	}
	if len(cfg.CustomClassifiers[0].Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(cfg.CustomClassifiers[0].Extensions))
	}
}
