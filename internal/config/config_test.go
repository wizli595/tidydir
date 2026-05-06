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
