package classifier

import (
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/scanner"
)

// Category represents what kind of thing an entry is.
type Category string

const (
	CatProject   Category = "project"
	CatDocument  Category = "document"
	CatMedia     Category = "media"
	CatFont      Category = "font"
	CatArchive   Category = "archive"
	CatDuplicate Category = "duplicate"
	CatJunk      Category = "junk"
	CatUnknown   Category = "unknown"
)

// Classification is the result of classifying an entry.
type Classification struct {
	Entry    scanner.Entry
	Category Category
	SubType  string // e.g. "go", "node", "spring" for projects
	Reason   string // human-readable explanation
}

// Classifier interface — each rule implements this.
type Classifier interface {
	Name() string
	Classify(entry scanner.Entry, allEntries []scanner.Entry) *Classification
}

// RunAll runs every classifier on every entry. First match wins.
func RunAll(classifiers []Classifier, entries []scanner.Entry) []Classification {
	var results []Classification

	for _, entry := range entries {
		var matched bool
		for _, c := range classifiers {
			if result := c.Classify(entry, entries); result != nil {
				results = append(results, *result)
				matched = true
				break
			}
		}
		if !matched {
			results = append(results, Classification{
				Entry:    entry,
				Category: CatUnknown,
				Reason:   "no rule matched",
			})
		}
	}

	return results
}

// DefaultClassifiers returns the standard set of classifiers in priority order.
// Custom classifiers from config run before the built-in filetype classifier.
func DefaultClassifiers(extraMarkers []config.ProjectMarker, customRules []config.ClassifierRule) []Classifier {
	classifiers := []Classifier{
		&JunkClassifier{},
		&DuplicateClassifier{},
		&ProjectClassifier{ExtraMarkers: extraMarkers},
	}

	if len(customRules) > 0 {
		classifiers = append(classifiers, &CustomClassifier{Rules: customRules})
	}

	classifiers = append(classifiers, &FileTypeClassifier{})
	return classifiers
}
