package classifier

import (
	"path/filepath"
	"strings"

	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/scanner"
)

// CustomClassifier classifies entries using user-defined rules from config.
type CustomClassifier struct {
	Rules []config.ClassifierRule
}

func (c *CustomClassifier) Name() string { return "custom" }

func (c *CustomClassifier) Classify(entry scanner.Entry, _ []scanner.Entry) *Classification {
	for _, rule := range c.Rules {
		if matchesRule(entry, rule) {
			return &Classification{
				Entry:    entry,
				Category: Category(rule.Category),
				SubType:  rule.SubType,
				Reason:   "custom classifier: " + rule.Name,
			}
		}
	}
	return nil
}

func matchesRule(entry scanner.Entry, rule config.ClassifierRule) bool {
	if len(rule.Extensions) > 0 {
		ext := strings.ToLower(entry.Ext)
		for _, e := range rule.Extensions {
			if ext == strings.ToLower(e) {
				return true
			}
		}
	}
	if len(rule.Patterns) > 0 {
		for _, p := range rule.Patterns {
			if matched, _ := filepath.Match(p, entry.Name); matched {
				return true
			}
		}
	}
	return false
}
