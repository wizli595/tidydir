package scanner

import (
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	Ext     string
	ModTime time.Time
}

type Options struct {
	Ignore []string
	Depth  int // 0 = top-level only, >0 = recurse N levels
}

func Scan(root string, opts Options) ([]Entry, error) {
	var entries []Entry
	return entries, scan(root, opts, 0, &entries)
}

func scan(dir string, opts Options, level int, entries *[]Entry) error {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, de := range dirEntries {
		name := de.Name()
		if ignored(name, opts.Ignore) {
			continue
		}

		info, err := de.Info()
		if err != nil {
			continue
		}

		entry := Entry{
			Name:    name,
			Path:    filepath.Join(dir, name),
			IsDir:   de.IsDir(),
			Size:    info.Size(),
			Ext:     filepath.Ext(name),
			ModTime: info.ModTime(),
		}
		*entries = append(*entries, entry)

		if de.IsDir() && opts.Depth > 0 && level < opts.Depth {
			scan(entry.Path, opts, level+1, entries)
		}
	}

	return nil
}

func ignored(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
	}
	return false
}
