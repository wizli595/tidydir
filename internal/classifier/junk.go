package classifier

import (
	"strings"

	"github.com/wizli595/tidydir/internal/scanner"
)

// JunkClassifier detects temp/system files that can be cleaned up.
type JunkClassifier struct{}

func (j *JunkClassifier) Name() string { return "junk" }

var junkNames = map[string]bool{
	".DS_Store":  true,
	"Thumbs.db":  true,
	".thumbs":    true,
	"__MACOSX":   true,
}

func (j *JunkClassifier) Classify(entry scanner.Entry, _ []scanner.Entry) *Classification {
	if junkNames[entry.Name] {
		return &Classification{
			Entry:    entry,
			Category: CatJunk,
			Reason:   "system junk file",
		}
	}

	// Detect temp/cache directories
	lower := strings.ToLower(entry.Name)
	if entry.IsDir && (lower == "node_modules" || lower == ".cache" || lower == "tmp") {
		return &Classification{
			Entry:    entry,
			Category: CatJunk,
			Reason:   "temp/cache directory",
		}
	}

	return nil
}
