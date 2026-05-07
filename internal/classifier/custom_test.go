package classifier

import (
	"testing"

	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/scanner"
)

func TestCustomClassifierByExtension(t *testing.T) {
	c := &CustomClassifier{
		Rules: []config.ClassifierRule{
			{Name: "data-files", Extensions: []string{".parquet", ".avro"}, Category: "document", SubType: "data"},
		},
	}

	entry := scanner.Entry{Name: "dataset.parquet", Path: "/tmp/dataset.parquet", Ext: ".parquet"}
	result := c.Classify(entry, nil)
	if result == nil {
		t.Fatal("expected classification, got nil")
	}
	if result.Category != CatDocument {
		t.Errorf("expected category 'document', got %q", result.Category)
	}
	if result.SubType != "data" {
		t.Errorf("expected subtype 'data', got %q", result.SubType)
	}
}

func TestCustomClassifierByPattern(t *testing.T) {
	c := &CustomClassifier{
		Rules: []config.ClassifierRule{
			{Name: "logs", Patterns: []string{"*.log", "*.log.*"}, Category: "junk"},
		},
	}

	entry := scanner.Entry{Name: "app.log", Path: "/tmp/app.log", Ext: ".log"}
	result := c.Classify(entry, nil)
	if result == nil {
		t.Fatal("expected classification, got nil")
	}
	if result.Category != CatJunk {
		t.Errorf("expected category 'junk', got %q", result.Category)
	}
}

func TestCustomClassifierNoMatch(t *testing.T) {
	c := &CustomClassifier{
		Rules: []config.ClassifierRule{
			{Name: "data-files", Extensions: []string{".parquet"}, Category: "document"},
		},
	}

	entry := scanner.Entry{Name: "readme.md", Path: "/tmp/readme.md", Ext: ".md"}
	result := c.Classify(entry, nil)
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestCustomClassifierExtensionCaseInsensitive(t *testing.T) {
	c := &CustomClassifier{
		Rules: []config.ClassifierRule{
			{Name: "images", Extensions: []string{".PNG"}, Category: "media"},
		},
	}

	entry := scanner.Entry{Name: "photo.png", Path: "/tmp/photo.png", Ext: ".png"}
	result := c.Classify(entry, nil)
	if result == nil {
		t.Fatal("expected classification for case-insensitive extension match")
	}
}
