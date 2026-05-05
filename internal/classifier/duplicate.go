package classifier

import (
	"strings"

	"github.com/wizli595/tidydir/internal/scanner"
)

// DuplicateClassifier detects redundant files:
// - zip files that have a matching extracted folder
// - files with " (1)", " - Copy" in the name
type DuplicateClassifier struct{}

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
	if !entry.IsDir && (strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".rar") || strings.HasSuffix(name, ".7z")) {
		baseName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".zip"), ".rar"), ".7z")
		for _, other := range allEntries {
			if other.IsDir && other.Name == baseName {
				return &Classification{
					Entry:    entry,
					Category: CatDuplicate,
					Reason:   "archive has matching folder: " + baseName + "/",
				}
			}
		}
	}

	return nil
}
