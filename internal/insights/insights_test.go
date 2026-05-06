package insights

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/scanner"
)

// --- NormalizeName ---

func TestNormalizeName_SpacesToHyphens(t *testing.T) {
	got := NormalizeName("My Document.pdf")
	if got != "my-document.pdf" {
		t.Errorf("got %q, want my-document.pdf", got)
	}
}

func TestNormalizeName_CamelCase(t *testing.T) {
	got := NormalizeName("CamelCase.txt")
	if got != "camel-case.txt" {
		t.Errorf("got %q, want camel-case.txt", got)
	}
}

func TestNormalizeName_UnderscoresToHyphens(t *testing.T) {
	got := NormalizeName("my_file_name.js")
	if got != "my-file-name.js" {
		t.Errorf("got %q, want my-file-name.js", got)
	}
}

func TestNormalizeName_MixedCase(t *testing.T) {
	got := NormalizeName("MyApp.go")
	if got != "my-app.go" {
		t.Errorf("got %q, want my-app.go", got)
	}
}

func TestNormalizeName_AlreadyClean(t *testing.T) {
	got := NormalizeName("clean-name.txt")
	if got != "clean-name.txt" {
		t.Errorf("got %q, want clean-name.txt (no change)", got)
	}
}

func TestNormalizeName_SkipsHidden(t *testing.T) {
	got := NormalizeName(".gitignore")
	if got != ".gitignore" {
		t.Errorf("got %q, want .gitignore (no change)", got)
	}
}

func TestNormalizeName_SkipsSpecial(t *testing.T) {
	specials := []string{"README.md", "LICENSE", "Makefile", "Dockerfile", "go.mod"}
	for _, name := range specials {
		got := NormalizeName(name)
		if got != name {
			t.Errorf("NormalizeName(%q) = %q, want no change", name, got)
		}
	}
}

func TestNormalizeName_Parentheses(t *testing.T) {
	got := NormalizeName("file (backup).txt")
	if got != "file-backup.txt" {
		t.Errorf("got %q, want file-backup.txt", got)
	}
}

func TestNormalizeName_ConsecutiveUppercase(t *testing.T) {
	got := NormalizeName("HTMLParser.go")
	if got != "html-parser.go" {
		t.Errorf("got %q, want html-parser.go", got)
	}
}

func TestNormalizeName_LowercaseExt(t *testing.T) {
	got := NormalizeName("File.TXT")
	if got != "file.txt" {
		t.Errorf("got %q, want file.txt", got)
	}
}

// --- FormatSize ---

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}

	for _, c := range cases {
		got := FormatSize(c.bytes)
		if got != c.want {
			t.Errorf("FormatSize(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}

// --- findLargeFiles ---

func TestFindLargeFiles(t *testing.T) {
	entries := []scanner.Entry{
		{Name: "small.txt", Size: 1024, IsDir: false},
		{Name: "big.bin", Size: 200 * 1024 * 1024, IsDir: false},
		{Name: "dir", Size: 0, IsDir: true},
	}

	results := findLargeFiles(entries, 100)
	if len(results) != 1 {
		t.Fatalf("expected 1 large file, got %d", len(results))
	}
	if results[0].Name != "big.bin" {
		t.Errorf("expected big.bin, got %s", results[0].Name)
	}
}

func TestFindLargeFiles_NoneAboveThreshold(t *testing.T) {
	entries := []scanner.Entry{
		{Name: "a.txt", Size: 50 * 1024 * 1024, IsDir: false},
	}

	results := findLargeFiles(entries, 100)
	if len(results) != 0 {
		t.Errorf("expected 0 large files, got %d", len(results))
	}
}

// --- findOldProjects ---

func TestFindOldProjects(t *testing.T) {
	old := time.Now().AddDate(0, -8, 0)
	recent := time.Now().AddDate(0, 0, -5)

	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "old-app", ModTime: old, IsDir: true},
			Category: classifier.CatProject,
		},
		{
			Entry:    scanner.Entry{Name: "new-app", ModTime: recent, IsDir: true},
			Category: classifier.CatProject,
		},
		{
			Entry:    scanner.Entry{Name: "doc.pdf", ModTime: old},
			Category: classifier.CatDocument,
		},
	}

	results := findOldProjects(classifications, 180)
	if len(results) != 1 {
		t.Fatalf("expected 1 old project, got %d", len(results))
	}
	if results[0].Name != "old-app" {
		t.Errorf("expected old-app, got %s", results[0].Name)
	}
}

// --- findNamingIssues ---

func TestFindNamingIssues(t *testing.T) {
	entries := []scanner.Entry{
		{Name: "Bad Name.txt", Path: "/tmp/Bad Name.txt"},
		{Name: "clean-name.txt", Path: "/tmp/clean-name.txt"},
		{Name: "folder", Path: "/tmp/folder", IsDir: true},
	}

	results := findNamingIssues(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 naming issue, got %d", len(results))
	}
	if results[0].Name != "Bad Name.txt" {
		t.Errorf("expected Bad Name.txt, got %s", results[0].Name)
	}
}

// --- findOrphanedDeps ---

func TestFindOrphanedDeps(t *testing.T) {
	dir := t.TempDir()
	nmDir := filepath.Join(dir, "node_modules")
	os.Mkdir(nmDir, 0755)
	// Create a file big enough to pass the 10MB threshold
	bigFile := filepath.Join(nmDir, "big.js")
	os.WriteFile(bigFile, make([]byte, 11*1024*1024), 0644)

	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: filepath.Base(dir), Path: dir, IsDir: true},
			Category: classifier.CatProject,
		},
	}

	results := findOrphanedDeps(classifications)
	if len(results) != 1 {
		t.Fatalf("expected 1 orphaned dep, got %d", len(results))
	}
	if results[0].Type != InsightOrphanedDeps {
		t.Errorf("got type %s, want orphaned_deps", results[0].Type)
	}
}

func TestFindOrphanedDeps_SkipsSmall(t *testing.T) {
	dir := t.TempDir()
	nmDir := filepath.Join(dir, "node_modules")
	os.Mkdir(nmDir, 0755)
	os.WriteFile(filepath.Join(nmDir, "tiny.js"), []byte("x"), 0644)

	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: filepath.Base(dir), Path: dir, IsDir: true},
			Category: classifier.CatProject,
		},
	}

	results := findOrphanedDeps(classifications)
	if len(results) != 0 {
		t.Errorf("expected 0 results for small dir, got %d", len(results))
	}
}

// --- Analyze integration ---

func TestAnalyze_ReturnsAllTypes(t *testing.T) {
	entries := []scanner.Entry{
		{Name: "huge.bin", Size: 200 * 1024 * 1024, IsDir: false},
		{Name: "Bad Name.txt", Path: "/tmp/Bad Name.txt"},
	}

	old := time.Now().AddDate(-1, 0, 0)
	classifications := []classifier.Classification{
		{
			Entry:    scanner.Entry{Name: "old-app", ModTime: old, IsDir: true},
			Category: classifier.CatProject,
		},
	}

	results := Analyze(entries, classifications, Options{LargeFileMB: 100, OldProjectDays: 180})
	if len(results) < 2 {
		t.Errorf("expected at least 2 insights, got %d", len(results))
	}

	types := map[InsightType]bool{}
	for _, r := range results {
		types[r.Type] = true
	}
	if !types[InsightLargeFile] {
		t.Error("expected large_file insight")
	}
	if !types[InsightNaming] {
		t.Error("expected naming insight")
	}
}
