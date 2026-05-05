package planner

import (
	"path/filepath"

	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/classifier"
)

// categoryFolders maps categories to their target subdirectory.
var categoryFolders = map[classifier.Category]string{
	classifier.CatProject:  "projects",
	classifier.CatDocument: "_docs",
	classifier.CatMedia:    "_media",
	classifier.CatFont:     "_fonts",
	classifier.CatArchive:  "_archives",
}

// Plan generates actions from classifications.
func Plan(classifications []classifier.Classification, rootPath string) []action.Action {
	var actions []action.Action

	for _, c := range classifications {
		switch c.Category {
		case classifier.CatJunk:
			actions = append(actions, action.Action{
				Type:   action.ActionDelete,
				Source: c.Entry.Path,
				Reason: c.Reason,
			})

		case classifier.CatDuplicate:
			actions = append(actions, action.Action{
				Type:   action.ActionDelete,
				Source: c.Entry.Path,
				Reason: c.Reason,
			})

		case classifier.CatProject:
			dest := filepath.Join(rootPath, "projects", c.SubType, c.Entry.Name)
			if dest != c.Entry.Path {
				actions = append(actions, action.Action{
					Type:   action.ActionMove,
					Source: c.Entry.Path,
					Dest:   dest,
					Reason: c.Reason + " → projects/" + c.SubType + "/",
				})
			}

		case classifier.CatDocument, classifier.CatMedia, classifier.CatFont, classifier.CatArchive:
			folder := categoryFolders[c.Category]
			dest := filepath.Join(rootPath, folder, c.Entry.Name)
			if dest != c.Entry.Path {
				actions = append(actions, action.Action{
					Type:   action.ActionMove,
					Source: c.Entry.Path,
					Dest:   dest,
					Reason: c.Reason + " → " + folder + "/",
				})
			}

		case classifier.CatUnknown:
			// Leave unknown files alone
		}
	}

	return actions
}
