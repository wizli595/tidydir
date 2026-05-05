package classifier

import (
	"strings"

	"github.com/wizli595/tidydir/internal/scanner"
)

// FileTypeClassifier categorizes files by their extension.
type FileTypeClassifier struct{}

func (f *FileTypeClassifier) Name() string { return "filetype" }

var extMap = map[string]Category{
	// Documents
	".pdf": CatDocument, ".doc": CatDocument, ".docx": CatDocument,
	".xls": CatDocument, ".xlsx": CatDocument, ".pptx": CatDocument,
	".txt": CatDocument, ".csv": CatDocument, ".html": CatDocument,

	// Media
	".png": CatMedia, ".jpg": CatMedia, ".jpeg": CatMedia,
	".gif": CatMedia, ".mp4": CatMedia, ".mp3": CatMedia,
	".mov": CatMedia, ".svg": CatMedia, ".webp": CatMedia,

	// Fonts
	".ttf": CatFont, ".otf": CatFont, ".woff": CatFont, ".woff2": CatFont,

	// Archives
	".zip": CatArchive, ".tar": CatArchive, ".gz": CatArchive,
	".rar": CatArchive, ".7z": CatArchive,
}

func (f *FileTypeClassifier) Classify(entry scanner.Entry, _ []scanner.Entry) *Classification {
	if entry.IsDir {
		return nil
	}

	ext := strings.ToLower(entry.Ext)
	if cat, ok := extMap[ext]; ok {
		return &Classification{
			Entry:    entry,
			Category: cat,
			Reason:   "extension " + ext,
		}
	}

	return nil
}
