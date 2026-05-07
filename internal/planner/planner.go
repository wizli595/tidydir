package planner

import (
	"path/filepath"
	"strings"

	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/insights"
	"github.com/wizli595/tidydir/internal/scanner"
)

func Plan(classifications []classifier.Classification, rootPath string, folders map[string]string, customRules []config.CustomRule) []action.Action {
	var actions []action.Action

	// Apply custom rules first
	customized := applyCustomRules(classifications, rootPath, customRules)
	actions = append(actions, customized...)

	// Track which entries were handled by custom rules
	handled := make(map[string]bool)
	for _, a := range customized {
		handled[a.Source] = true
	}

	for _, c := range classifications {
		if handled[c.Entry.Path] {
			continue
		}

		switch c.Category {
		case classifier.CatJunk, classifier.CatDuplicate:
			actions = append(actions, action.Action{
				Type:   action.ActionDelete,
				Source: c.Entry.Path,
				Reason: c.Reason,
			})

		case classifier.CatProject:
			folder := folders["project"]
			if strings.Contains(c.SubType, "..") || strings.Contains(c.Entry.Name, "..") {
				continue
			}
			dest := filepath.Join(rootPath, folder, c.SubType, c.Entry.Name)
			if dest != c.Entry.Path {
				actions = append(actions, action.Action{
					Type:   action.ActionMove,
					Source: c.Entry.Path,
					Dest:   dest,
					Reason: c.Reason + " -> " + folder + "/" + c.SubType + "/",
				})
			}

		case classifier.CatDocument, classifier.CatMedia, classifier.CatFont, classifier.CatArchive:
			folder := folders[string(c.Category)]
			if folder == "" {
				continue
			}
			dest := filepath.Join(rootPath, folder, c.Entry.Name)
			if dest != c.Entry.Path {
				actions = append(actions, action.Action{
					Type:   action.ActionMove,
					Source: c.Entry.Path,
					Dest:   dest,
					Reason: c.Reason + " -> " + folder + "/",
				})
			}
		}
	}

	return actions
}

// PlanRenames generates rename actions for files with naming issues.
// Skips entries already handled by other actions and directories.
func PlanRenames(entries []scanner.Entry, rootPath string, existingActions []action.Action) []action.Action {
	handled := make(map[string]bool)
	for _, a := range existingActions {
		handled[a.Source] = true
	}

	var renames []action.Action
	for _, e := range entries {
		if e.IsDir || handled[e.Path] {
			continue
		}
		// Only rename top-level files
		if filepath.Dir(e.Path) != rootPath {
			continue
		}
		normalized := insights.NormalizeName(e.Name)
		if normalized != e.Name {
			renames = append(renames, action.Action{
				Type:   action.ActionRename,
				Source: e.Path,
				Dest:   filepath.Join(rootPath, normalized),
				Reason: "normalize name",
			})
		}
	}
	return renames
}

func applyCustomRules(classifications []classifier.Classification, rootPath string, rules []config.CustomRule) []action.Action {
	var actions []action.Action

	for _, c := range classifications {
		for _, rule := range rules {
			if matched, _ := filepath.Match(rule.Pattern, c.Entry.Name); matched {
				dest := filepath.Join(rootPath, rule.Dest, c.Entry.Name)
				if dest != c.Entry.Path {
					actions = append(actions, action.Action{
						Type:   action.ActionMove,
						Source: c.Entry.Path,
						Dest:   dest,
						Reason: "custom rule: " + rule.Pattern + " -> " + rule.Dest + "/",
					})
				}
				break
			}
		}
	}

	return actions
}
