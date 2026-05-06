package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanTopLevel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	entries, err := Scan(dir, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name] = true
	}
	if !names["file.txt"] || !names["subdir"] {
		t.Errorf("expected file.txt and subdir, got %v", names)
	}
}

func TestScanEntryFields(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.pdf"), []byte("data"), 0644)

	entries, _ := Scan(dir, Options{})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.Name != "doc.pdf" {
		t.Errorf("name = %q, want doc.pdf", e.Name)
	}
	if e.IsDir {
		t.Error("expected IsDir = false")
	}
	if e.Ext != ".pdf" {
		t.Errorf("ext = %q, want .pdf", e.Ext)
	}
	if e.Size != 4 {
		t.Errorf("size = %d, want 4", e.Size)
	}
	if e.ModTime.IsZero() {
		t.Error("expected non-zero ModTime")
	}
}

func TestScanIgnorePatterns(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "skip.log"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "desktop.ini"), []byte(""), 0644)

	entries, _ := Scan(dir, Options{Ignore: []string{"*.log", "desktop.ini"}})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "keep.txt" {
		t.Errorf("expected keep.txt, got %s", entries[0].Name)
	}
}

func TestScanDepthZero(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "nested.txt"), []byte(""), 0644)

	entries, _ := Scan(dir, Options{Depth: 0})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (just sub/), got %d", len(entries))
	}
	if entries[0].Name != "sub" {
		t.Errorf("expected sub, got %s", entries[0].Name)
	}
}

func TestScanDepthRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "nested.txt"), []byte(""), 0644)

	entries, _ := Scan(dir, Options{Depth: 1})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name] = true
	}
	if !names["sub"] || !names["nested.txt"] {
		t.Errorf("expected sub and nested.txt, got %v", names)
	}
}

func TestScanEmptyDir(t *testing.T) {
	dir := t.TempDir()
	entries, err := Scan(dir, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
