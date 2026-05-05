package classifier

import (
	"os"
	"path/filepath"

	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/scanner"
)

type ProjectClassifier struct {
	ExtraMarkers []config.ProjectMarker
}

func (p *ProjectClassifier) Name() string { return "project" }

var builtinMarkers = map[string]string{
	"go.mod":           "go",
	"package.json":     "node",
	"Cargo.toml":       "rust",
	"pom.xml":          "java",
	"build.gradle":     "java",
	"pubspec.yaml":     "flutter",
	"requirements.txt": "python",
	"pyproject.toml":   "python",
	"Gemfile":          "ruby",
	"*.csproj":         "dotnet",
	"*.sln":            "dotnet",
}

func (p *ProjectClassifier) Classify(entry scanner.Entry, _ []scanner.Entry) *Classification {
	if !entry.IsDir {
		return nil
	}

	// Check built-in markers
	if result := checkMarkers(entry, builtinMarkers); result != nil {
		return result
	}

	// Check user-defined markers from config
	for _, m := range p.ExtraMarkers {
		if _, err := os.Stat(filepath.Join(entry.Path, m.File)); err == nil {
			return &Classification{
				Entry:    entry,
				Category: CatProject,
				SubType:  m.Type,
				Reason:   "found " + m.File,
			}
		}
	}

	return nil
}

func checkMarkers(entry scanner.Entry, markers map[string]string) *Classification {
	for marker, subtype := range markers {
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
