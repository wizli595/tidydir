package insights

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/scanner"
)

type InsightType string

const (
	InsightLargeFile    InsightType = "large_file"
	InsightOldProject   InsightType = "old_project"
	InsightDirtyGit     InsightType = "dirty_git"
	InsightOrphanedDeps InsightType = "orphaned_deps"
	InsightNaming       InsightType = "naming"
)

type Insight struct {
	Type    InsightType
	Path    string
	Name    string
	Message string
}

type Options struct {
	LargeFileMB    int
	OldProjectDays int
}

func Analyze(entries []scanner.Entry, classifications []classifier.Classification, opts Options) []Insight {
	if opts.LargeFileMB <= 0 {
		opts.LargeFileMB = 100
	}
	if opts.OldProjectDays <= 0 {
		opts.OldProjectDays = 180
	}

	var all []Insight
	all = append(all, findLargeFiles(entries, opts.LargeFileMB)...)
	all = append(all, findOldProjects(classifications, opts.OldProjectDays)...)
	all = append(all, findDirtyGitRepos(classifications)...)
	all = append(all, findOrphanedDeps(classifications)...)
	all = append(all, findNamingIssues(entries)...)
	return all
}

// --- Large files ---

func findLargeFiles(entries []scanner.Entry, thresholdMB int) []Insight {
	threshold := int64(thresholdMB) * 1024 * 1024
	var out []Insight
	for _, e := range entries {
		if !e.IsDir && e.Size >= threshold {
			out = append(out, Insight{
				Type:    InsightLargeFile,
				Path:    e.Path,
				Name:    e.Name,
				Message: FormatSize(e.Size),
			})
		}
	}
	return out
}

// --- Old projects ---

func findOldProjects(classifications []classifier.Classification, days int) []Insight {
	cutoff := time.Now().AddDate(0, 0, -days)
	var out []Insight
	for _, c := range classifications {
		if c.Category != classifier.CatProject {
			continue
		}
		if c.Entry.ModTime.Before(cutoff) {
			months := int(time.Since(c.Entry.ModTime).Hours() / 24 / 30)
			msg := fmt.Sprintf("last modified %d months ago", months)
			if months <= 1 {
				msg = "last modified over " + fmt.Sprintf("%d", days) + " days ago"
			}
			out = append(out, Insight{
				Type:    InsightOldProject,
				Path:    c.Entry.Path,
				Name:    c.Entry.Name,
				Message: msg,
			})
		}
	}
	return out
}

// --- Dirty git repos ---

func findDirtyGitRepos(classifications []classifier.Classification) []Insight {
	var out []Insight
	for _, c := range classifications {
		if c.Category != classifier.CatProject {
			continue
		}
		if ins := checkGitStatus(c.Entry); ins != nil {
			out = append(out, *ins)
		}
	}
	return out
}

func checkGitStatus(entry scanner.Entry) *Insight {
	if _, err := os.Stat(filepath.Join(entry.Path, ".git")); err != nil {
		return nil
	}
	raw, err := exec.Command("git", "-C", entry.Path, "status", "--porcelain").Output()
	if err != nil || strings.TrimSpace(string(raw)) == "" {
		return nil
	}
	n := len(strings.Split(strings.TrimSpace(string(raw)), "\n"))
	return &Insight{
		Type:    InsightDirtyGit,
		Path:    entry.Path,
		Name:    entry.Name,
		Message: fmt.Sprintf("%d uncommitted changes", n),
	}
}

// --- Orphaned / heavy dependency dirs ---

var heavyDirNames = []string{
	"node_modules", "vendor", "build", "dist", "target",
	".cache", "__pycache__", ".next", ".nuxt",
}

func findOrphanedDeps(classifications []classifier.Classification) []Insight {
	var out []Insight
	for _, c := range classifications {
		if c.Category != classifier.CatProject {
			continue
		}
		for _, name := range heavyDirNames {
			p := filepath.Join(c.Entry.Path, name)
			info, err := os.Stat(p)
			if err != nil || !info.IsDir() {
				continue
			}
			size := dirSize(p)
			if size < 10*1024*1024 { // skip if under 10 MB
				continue
			}
			out = append(out, Insight{
				Type:    InsightOrphanedDeps,
				Path:    p,
				Name:    c.Entry.Name + "/" + name,
				Message: FormatSize(size),
			})
		}
	}
	return out
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// --- Naming issues ---

func findNamingIssues(entries []scanner.Entry) []Insight {
	var out []Insight
	for _, e := range entries {
		if e.IsDir {
			continue
		}
		suggested := NormalizeName(e.Name)
		if suggested != e.Name {
			out = append(out, Insight{
				Type:    InsightNaming,
				Path:    e.Path,
				Name:    e.Name,
				Message: suggested,
			})
		}
	}
	return out
}

var specialFiles = map[string]bool{
	"README.md": true, "README": true, "LICENSE": true, "LICENSE.md": true,
	"Makefile": true, "Dockerfile": true, "CHANGELOG.md": true,
	"CONTRIBUTING.md": true, "FEATURES.md": true,
	"go.mod": true, "go.sum": true, "package.json": true,
	"Cargo.toml": true, "Gemfile": true, "Rakefile": true,
	"CMakeLists.txt": true, "pom.xml": true, "build.gradle": true,
}

// NormalizeName converts a filename to kebab-case.
// Returns the original name if no change is needed.
func NormalizeName(name string) string {
	if strings.HasPrefix(name, ".") {
		return name
	}
	if specialFiles[name] {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	if isClean(base) {
		return name
	}

	normalized := toKebab(base)
	if normalized == "" {
		return name
	}
	return normalized + strings.ToLower(ext)
}

func isClean(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) || r == ' ' || r == '_' || r == '(' || r == ')' {
			return false
		}
	}
	return true
}

func toKebab(s string) string {
	var buf strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		switch {
		case r == ' ' || r == '_' || r == '(' || r == ')':
			if buf.Len() > 0 {
				last := buf.String()
				if last[len(last)-1] != '-' {
					buf.WriteRune('-')
				}
			}
		case unicode.IsUpper(r):
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					buf.WriteRune('-')
				} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					buf.WriteRune('-')
				}
			}
			buf.WriteRune(unicode.ToLower(r))
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.':
			buf.WriteRune(r)
		default:
			// skip other special chars
		}
	}

	result := buf.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")
	return result
}

// FormatSize formats bytes into a human-readable string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
