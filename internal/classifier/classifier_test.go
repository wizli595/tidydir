package classifier

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/scanner"
)

// --- JunkClassifier ---

func TestJunkClassifier_DetectsJunk(t *testing.T) {
	c := &JunkClassifier{}
	cases := []string{".DS_Store", "Thumbs.db", "__MACOSX"}

	for _, name := range cases {
		entry := scanner.Entry{Name: name, Path: "/tmp/" + name}
		result := c.Classify(entry, nil)
		if result == nil {
			t.Errorf("expected %s to be classified as junk", name)
		} else if result.Category != CatJunk {
			t.Errorf("%s: got category %s, want junk", name, result.Category)
		}
	}
}

func TestJunkClassifier_DetectsCacheDirs(t *testing.T) {
	c := &JunkClassifier{}
	dirs := []string{"node_modules", ".cache", "tmp"}

	for _, name := range dirs {
		entry := scanner.Entry{Name: name, Path: "/tmp/" + name, IsDir: true}
		result := c.Classify(entry, nil)
		if result == nil {
			t.Errorf("expected %s to be classified as junk", name)
		}
	}
}

func TestJunkClassifier_IgnoresNormal(t *testing.T) {
	c := &JunkClassifier{}
	entry := scanner.Entry{Name: "readme.md", Path: "/tmp/readme.md"}
	if c.Classify(entry, nil) != nil {
		t.Error("readme.md should not be classified as junk")
	}
}

// --- DuplicateClassifier ---

func TestDuplicateClassifier_CopyPattern(t *testing.T) {
	c := &DuplicateClassifier{}
	cases := []string{"file (1).txt", "doc (2).pdf", "thing - Copy.zip"}

	for _, name := range cases {
		entry := scanner.Entry{Name: name, Path: "/tmp/" + name}
		result := c.Classify(entry, nil)
		if result == nil {
			t.Errorf("expected %s to be duplicate", name)
		} else if result.Category != CatDuplicate {
			t.Errorf("%s: got %s, want duplicate", name, result.Category)
		}
	}
}

func TestDuplicateClassifier_ZipWithFolder(t *testing.T) {
	c := &DuplicateClassifier{}
	entries := []scanner.Entry{
		{Name: "project.zip", Path: "/tmp/project.zip", IsDir: false},
		{Name: "project", Path: "/tmp/project", IsDir: true},
	}

	result := c.Classify(entries[0], entries)
	if result == nil {
		t.Fatal("expected zip with matching folder to be duplicate")
	}
	if result.Category != CatDuplicate {
		t.Errorf("got %s, want duplicate", result.Category)
	}
}

func TestDuplicateClassifier_ZipWithoutFolder(t *testing.T) {
	c := &DuplicateClassifier{}
	entries := []scanner.Entry{
		{Name: "archive.zip", Path: "/tmp/archive.zip", IsDir: false},
	}

	result := c.Classify(entries[0], entries)
	if result != nil {
		t.Error("zip without matching folder should not be duplicate")
	}
}

func TestDuplicateClassifier_IgnoresNormal(t *testing.T) {
	c := &DuplicateClassifier{}
	entry := scanner.Entry{Name: "report.pdf", Path: "/tmp/report.pdf"}
	if c.Classify(entry, nil) != nil {
		t.Error("report.pdf should not be duplicate")
	}
}

// --- ProjectClassifier ---

func TestProjectClassifier_DetectsGoProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	c := &ProjectClassifier{}
	entry := scanner.Entry{Name: filepath.Base(dir), Path: dir, IsDir: true}
	result := c.Classify(entry, nil)

	if result == nil {
		t.Fatal("expected go project to be detected")
	}
	if result.Category != CatProject {
		t.Errorf("got %s, want project", result.Category)
	}
	if result.SubType != "go" {
		t.Errorf("got subtype %s, want go", result.SubType)
	}
}

func TestProjectClassifier_DetectsNodeProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)

	c := &ProjectClassifier{}
	entry := scanner.Entry{Name: filepath.Base(dir), Path: dir, IsDir: true}
	result := c.Classify(entry, nil)

	if result == nil {
		t.Fatal("expected node project to be detected")
	}
	if result.SubType != "node" {
		t.Errorf("got subtype %s, want node", result.SubType)
	}
}

func TestProjectClassifier_ExtraMarkers(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(""), 0644)

	c := &ProjectClassifier{
		ExtraMarkers: []config.ProjectMarker{{File: "docker-compose.yml", Type: "docker"}},
	}
	entry := scanner.Entry{Name: filepath.Base(dir), Path: dir, IsDir: true}
	result := c.Classify(entry, nil)

	if result == nil {
		t.Fatal("expected docker project to be detected")
	}
	if result.SubType != "docker" {
		t.Errorf("got subtype %s, want docker", result.SubType)
	}
}

func TestProjectClassifier_IgnoresFiles(t *testing.T) {
	c := &ProjectClassifier{}
	entry := scanner.Entry{Name: "file.go", Path: "/tmp/file.go", IsDir: false}
	if c.Classify(entry, nil) != nil {
		t.Error("files should not be classified as projects")
	}
}

func TestProjectClassifier_IgnoresEmptyDir(t *testing.T) {
	dir := t.TempDir()
	c := &ProjectClassifier{}
	entry := scanner.Entry{Name: filepath.Base(dir), Path: dir, IsDir: true}
	if c.Classify(entry, nil) != nil {
		t.Error("empty dir should not be classified as project")
	}
}

// --- FileTypeClassifier ---

func TestFileTypeClassifier_Documents(t *testing.T) {
	c := &FileTypeClassifier{}
	exts := []string{".pdf", ".docx", ".xlsx", ".csv", ".txt"}

	for _, ext := range exts {
		entry := scanner.Entry{Name: "file" + ext, Path: "/tmp/file" + ext, Ext: ext}
		result := c.Classify(entry, nil)
		if result == nil {
			t.Errorf("expected %s to be classified", ext)
		} else if result.Category != CatDocument {
			t.Errorf("%s: got %s, want document", ext, result.Category)
		}
	}
}

func TestFileTypeClassifier_Media(t *testing.T) {
	c := &FileTypeClassifier{}
	exts := []string{".png", ".jpg", ".mp4", ".mp3", ".svg"}

	for _, ext := range exts {
		entry := scanner.Entry{Name: "file" + ext, Path: "/tmp/file" + ext, Ext: ext}
		result := c.Classify(entry, nil)
		if result == nil {
			t.Errorf("expected %s to be classified", ext)
		} else if result.Category != CatMedia {
			t.Errorf("%s: got %s, want media", ext, result.Category)
		}
	}
}

func TestFileTypeClassifier_Fonts(t *testing.T) {
	c := &FileTypeClassifier{}
	entry := scanner.Entry{Name: "font.ttf", Path: "/tmp/font.ttf", Ext: ".ttf"}
	result := c.Classify(entry, nil)
	if result == nil || result.Category != CatFont {
		t.Error("expected .ttf to be classified as font")
	}
}

func TestFileTypeClassifier_Archives(t *testing.T) {
	c := &FileTypeClassifier{}
	entry := scanner.Entry{Name: "data.zip", Path: "/tmp/data.zip", Ext: ".zip"}
	result := c.Classify(entry, nil)
	if result == nil || result.Category != CatArchive {
		t.Error("expected .zip to be classified as archive")
	}
}

func TestFileTypeClassifier_IgnoresDirs(t *testing.T) {
	c := &FileTypeClassifier{}
	entry := scanner.Entry{Name: "media", Path: "/tmp/media", IsDir: true, Ext: ""}
	if c.Classify(entry, nil) != nil {
		t.Error("directories should not be classified by filetype")
	}
}

func TestFileTypeClassifier_UnknownExt(t *testing.T) {
	c := &FileTypeClassifier{}
	entry := scanner.Entry{Name: "data.xyz", Path: "/tmp/data.xyz", Ext: ".xyz"}
	if c.Classify(entry, nil) != nil {
		t.Error("unknown extension should return nil")
	}
}

// --- RunAll ---

func TestRunAll_FirstMatchWins(t *testing.T) {
	entries := []scanner.Entry{
		{Name: ".DS_Store", Path: "/tmp/.DS_Store"},
	}

	classifiers := DefaultClassifiers(nil)
	results := RunAll(classifiers, entries)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// JunkClassifier has higher priority than FileType
	if results[0].Category != CatJunk {
		t.Errorf("got %s, want junk (first match wins)", results[0].Category)
	}
}

func TestRunAll_UnmatchedBecomesUnknown(t *testing.T) {
	entries := []scanner.Entry{
		{Name: "random.xyz", Path: "/tmp/random.xyz", Ext: ".xyz"},
	}

	results := RunAll(DefaultClassifiers(nil), entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Category != CatUnknown {
		t.Errorf("got %s, want unknown", results[0].Category)
	}
}

func TestDefaultClassifiers_Order(t *testing.T) {
	classifiers := DefaultClassifiers(nil)
	if len(classifiers) != 4 {
		t.Fatalf("expected 4 classifiers, got %d", len(classifiers))
	}
	if classifiers[0].Name() != "junk" {
		t.Errorf("first classifier should be junk, got %s", classifiers[0].Name())
	}
	if classifiers[1].Name() != "duplicate" {
		t.Errorf("second should be duplicate, got %s", classifiers[1].Name())
	}
	if classifiers[2].Name() != "project" {
		t.Errorf("third should be project, got %s", classifiers[2].Name())
	}
	if classifiers[3].Name() != "filetype" {
		t.Errorf("fourth should be filetype, got %s", classifiers[3].Name())
	}
}
