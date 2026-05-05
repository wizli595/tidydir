package scanner

import (
	"os"
	"path/filepath"
)

// Entry represents a file or directory in the scanned path.
type Entry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
	Ext   string // lowercase, e.g. ".pdf"
}

// Scan reads the top-level contents of a directory (non-recursive).
func Scan(root string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, de := range dirEntries {
		info, err := de.Info()
		if err != nil {
			continue
		}

		name := de.Name()
		if name == "desktop.ini" || name == "Thumbs.db" {
			continue
		}

		entries = append(entries, Entry{
			Name:  name,
			Path:  filepath.Join(root, name),
			IsDir: de.IsDir(),
			Size:  info.Size(),
			Ext:   filepath.Ext(name),
		})
	}

	return entries, nil
}
