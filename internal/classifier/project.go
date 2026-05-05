package classifier

import (
	"os"
	"path/filepath"

	"github.com/wizli595/tidydir/internal/scanner"
)

// ProjectClassifier detects dev project directories by marker files.
type ProjectClassifier struct{}

func (p *ProjectClassifier) Name() string { return "project" }

// Project markers: file to look for → subtype label.
var projectMarkers = map[string]string{
	"go.mod":         "go",
	"package.json":   "node",
	"Cargo.toml":     "rust",
	"pom.xml":        "java",
	"build.gradle":   "java",
	"pubspec.yaml":   "flutter",
	"requirements.txt": "python",
	"pyproject.toml": "python",
	"Gemfile":        "ruby",
	"*.csproj":       "dotnet",
	"*.sln":          "dotnet",
}

func (p *ProjectClassifier) Classify(entry scanner.Entry, _ []scanner.Entry) *Classification {
	if !entry.IsDir {
		return nil
	}

	for marker, subtype := range projectMarkers {
		// Handle glob patterns like *.csproj
		if marker[0] == '*' {
			matches, _ := filepath.Glob(filepath.Join(entry.Path, marker))
			if len(matches) > 0 {
				return &Classification{
					Entry:    entry,
					Category: CatProject,
					SubType:  subtype,
					Reason:   "found " + marker,
				}
			}
		} else {
			if _, err := os.Stat(filepath.Join(entry.Path, marker)); err == nil {
				return &Classification{
					Entry:    entry,
					Category: CatProject,
					SubType:  subtype,
					Reason:   "found " + marker,
				}
			}
		}
	}

	return nil
}
