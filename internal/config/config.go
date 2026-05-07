package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Folders           map[string]string  `yaml:"folders"`
	ProjectMarkers    []ProjectMarker    `yaml:"project_markers"`
	Ignore            []string           `yaml:"ignore"`
	CustomRules       []CustomRule       `yaml:"custom_rules"`
	CustomClassifiers []ClassifierRule   `yaml:"custom_classifiers"`
	Profiles          map[string]Profile `yaml:"profiles"`
	LargeFileMB       int                `yaml:"large_file_mb"`
	OldProjectDays    int                `yaml:"old_project_days"`
}

type ProjectMarker struct {
	File string `yaml:"file"`
	Type string `yaml:"type"`
}

type CustomRule struct {
	Pattern string `yaml:"pattern"`
	Dest    string `yaml:"dest"`
}

// ClassifierRule defines a user-configured classifier (plugin system).
type ClassifierRule struct {
	Name       string   `yaml:"name"`
	Extensions []string `yaml:"extensions"`
	Patterns   []string `yaml:"patterns"`
	Category   string   `yaml:"category"`
	SubType    string   `yaml:"subtype"`
}

// Profile holds alternate settings for different use cases.
type Profile struct {
	Folders           map[string]string `yaml:"folders"`
	Ignore            []string          `yaml:"ignore"`
	CustomRules       []CustomRule      `yaml:"custom_rules"`
	CustomClassifiers []ClassifierRule  `yaml:"custom_classifiers"`
	LargeFileMB       int               `yaml:"large_file_mb"`
	OldProjectDays    int               `yaml:"old_project_days"`
}

func Load(targetDir string) *Config {
	return LoadWithProfile(targetDir, "")
}

func LoadWithProfile(targetDir, profile string) *Config {
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

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	yaml.Unmarshal(data, cfg)

	if profile != "" {
		applyProfile(cfg, profile)
	}

	return cfg
}

func applyProfile(cfg *Config, name string) {
	p, ok := cfg.Profiles[name]
	if !ok {
		return
	}
	if p.Folders != nil {
		for k, v := range p.Folders {
			cfg.Folders[k] = v
		}
	}
	if p.Ignore != nil {
		cfg.Ignore = p.Ignore
	}
	if p.CustomRules != nil {
		cfg.CustomRules = p.CustomRules
	}
	if p.CustomClassifiers != nil {
		cfg.CustomClassifiers = p.CustomClassifiers
	}
	if p.LargeFileMB > 0 {
		cfg.LargeFileMB = p.LargeFileMB
	}
	if p.OldProjectDays > 0 {
		cfg.OldProjectDays = p.OldProjectDays
	}
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
