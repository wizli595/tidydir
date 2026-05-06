package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Folders        map[string]string `yaml:"folders"`
	ProjectMarkers []ProjectMarker   `yaml:"project_markers"`
	Ignore         []string          `yaml:"ignore"`
	CustomRules    []CustomRule      `yaml:"custom_rules"`
	LargeFileMB    int               `yaml:"large_file_mb"`
	OldProjectDays int               `yaml:"old_project_days"`
}

type ProjectMarker struct {
	File string `yaml:"file"`
	Type string `yaml:"type"`
}

type CustomRule struct {
	Pattern string `yaml:"pattern"`
	Dest    string `yaml:"dest"`
}

func Load(targetDir string) *Config {
	cfg := &Config{
		Folders: map[string]string{
			"project":  "projects",
			"document": "_docs",
			"media":    "_media",
			"font":     "_fonts",
			"archive":  "_archives",
		},
		Ignore:         []string{"desktop.ini", "*.lnk", "My Music", "My Pictures", "My Videos"},
		LargeFileMB:    100,
		OldProjectDays: 180,
	}

	path := findFile(targetDir)
	if path == "" {
		return cfg
	}

	data, _ := os.ReadFile(path)
	yaml.Unmarshal(data, cfg)
	return cfg
}

func findFile(targetDir string) string {
	candidates := []string{
		filepath.Join(targetDir, "tidydir.yaml"),
		filepath.Join(targetDir, ".tidydir.yaml"),
	}

	exe, err := os.Executable()
	if err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "config", "rules.yaml"))
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
