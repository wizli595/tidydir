package planner

import (
	"path/filepath"

	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
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
