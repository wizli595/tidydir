package classifier

import (
	"strings"

	"github.com/wizli595/tidydir/internal/scanner"
)

// DuplicateClassifier detects redundant files:
// - zip files that have a matching extracted folder
// - files with " (1)", " - Copy" in the name
type DuplicateClassifier struct {
	dirIndex map[string]bool // cached directory name lookup, built on first use
}

func (d *DuplicateClassifier) Name() string { return "duplicate" }

func (d *DuplicateClassifier) Classify(entry scanner.Entry, allEntries []scanner.Entry) *Classification {
	name := entry.Name

	// Check for "(1)", "(2)", "- Copy" patterns
	if strings.Contains(name, "(1)") || strings.Contains(name, "(2)") || strings.Contains(name, "- Copy") {
		return &Classification{
			Entry:    entry,
			Category: CatDuplicate,
			Reason:   "looks like a copy",
		}
	}

	// Check for zip files with a matching extracted folder
	if !entry.IsDir {
		for _, ext := range archiveExts {
			if strings.HasSuffix(name, ext) {
				d.ensureDirIndex(allEntries)
				baseName := strings.TrimSuffix(name, ext)
				if d.dirIndex[baseName] {
					return &Classification{
						Entry:    entry,
						Category: CatDuplicate,
						Reason:   "archive has matching folder: " + baseName + "/",
					}
				}
				break
			}
		}
	}

	return nil
}

var archiveExts = []string{".zip", ".rar", ".7z"}

// ensureDirIndex builds the directory name lookup map once per run.
func (d *DuplicateClassifier) ensureDirIndex(entries []scanner.Entry) {
	if d.dirIndex != nil {
		return
	}
	d.dirIndex = make(map[string]bool, len(entries)/2)
	for _, e := range entries {
		if e.IsDir {
			d.dirIndex[e.Name] = true
		}
	}
}
