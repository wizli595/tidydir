package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/insights"
)

var (
	titleStyle   lipgloss.Style
	moveStyle    lipgloss.Style
	deleteStyle  lipgloss.Style
	renameStyle  lipgloss.Style
	dimStyle     lipgloss.Style
	pathStyle    lipgloss.Style
	countStyle   lipgloss.Style
	promptStyle  lipgloss.Style
	warnStyle    lipgloss.Style
	dangerStyle  lipgloss.Style
	infoStyle    lipgloss.Style
	suggestStyle lipgloss.Style
)

func init() {
	dark := termenv.HasDarkBackground()

	if dark {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
		dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	} else {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0"))
		dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
		infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	}

	// These work on both dark and light
	moveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	deleteStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	renameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	pathStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	countStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	promptStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	warnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	dangerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	suggestStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
}

func ShowInsights(items []insights.Insight) {
	if len(items) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(titleStyle.Render("  INSIGHTS"))
	fmt.Println(dimStyle.Render("  " + strings.Repeat("-", 56)))

	large := filterInsights(items, insights.InsightLargeFile)
	dirty := filterInsights(items, insights.InsightDirtyGit)
	old := filterInsights(items, insights.InsightOldProject)
	heavy := filterInsights(items, insights.InsightOrphanedDeps)
	naming := filterInsights(items, insights.InsightNaming)

	if len(large) > 0 {
		fmt.Println(warnStyle.Render("  LARGE FILES"))
		for i, ins := range large {
			fmt.Printf("    %s %s  %s\n",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				pathStyle.Render(ins.Name),
				warnStyle.Render(ins.Message))
		}
		fmt.Println()
	}

	if len(dirty) > 0 {
		fmt.Println(dangerStyle.Render("  DIRTY GIT REPOS"))
		for i, ins := range dirty {
			fmt.Printf("    %s %s  %s\n",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				pathStyle.Render(ins.Name+"/"),
				dangerStyle.Render(ins.Message))
		}
		fmt.Println()
	}

	if len(old) > 0 {
		fmt.Println(infoStyle.Render("  OLD PROJECTS"))
		for i, ins := range old {
			fmt.Printf("    %s %s  %s\n",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				pathStyle.Render(ins.Name+"/"),
				infoStyle.Render(ins.Message))
		}
		fmt.Println()
	}

	if len(heavy) > 0 {
		fmt.Println(warnStyle.Render("  HEAVY DIRECTORIES"))
		for i, ins := range heavy {
			fmt.Printf("    %s %s  %s\n",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				pathStyle.Render(ins.Name),
				warnStyle.Render(ins.Message))
		}
		fmt.Println()
	}

	if len(naming) > 0 {
		fmt.Println(suggestStyle.Render("  NAMING"))
		for i, ins := range naming {
			fmt.Printf("    %s %s %s %s\n",
				dimStyle.Render(fmt.Sprintf("%2d.", i+1)),
				pathStyle.Render(ins.Name),
				dimStyle.Render("->"),
				suggestStyle.Render(ins.Message))
		}
		fmt.Println()
	}

	fmt.Println(dimStyle.Render("  " + strings.Repeat("-", 56)))
}

func filterInsights(items []insights.Insight, t insights.InsightType) []insights.Insight {
	var out []insights.Insight
	for _, ins := range items {
		if ins.Type == t {
			out = append(out, ins)
		}
	}
	return out
}

func ShowPlan(actions []action.Action) {
	moves, deletes, renames := countActions(actions)

	fmt.Println()
	fmt.Println(titleStyle.Render("  PLAN"))
	fmt.Println(dimStyle.Render("  " + strings.Repeat("─", 56)))

	// Summary line
	parts := []string{}
	if moves > 0 {
		parts = append(parts, moveStyle.Render(fmt.Sprintf("%d moves", moves)))
	}
	if deletes > 0 {
		parts = append(parts, deleteStyle.Render(fmt.Sprintf("%d deletes", deletes)))
	}
	if renames > 0 {
		parts = append(parts, renameStyle.Render(fmt.Sprintf("%d renames", renames)))
	}
	fmt.Printf("  %s: %s\n\n", countStyle.Render(fmt.Sprintf("%d actions", len(actions))), strings.Join(parts, dimStyle.Render(" | ")))

	// Group by type
	if moves > 0 {
		fmt.Println(moveStyle.Render("  MOVE"))
		printActions(actions, action.ActionMove)
		fmt.Println()
	}
	if deletes > 0 {
		fmt.Println(deleteStyle.Render("  DELETE"))
		printActions(actions, action.ActionDelete)
		fmt.Println()
	}
	if renames > 0 {
		fmt.Println(renameStyle.Render("  RENAME"))
		printActions(actions, action.ActionRename)
		fmt.Println()
	}

	fmt.Println(dimStyle.Render("  " + strings.Repeat("─", 56)))
}

func printActions(actions []action.Action, filterType action.ActionType) {
	idx := 1
	for _, a := range actions {
		if a.Type != filterType {
			continue
		}
		name := filepath.Base(a.Source)
		switch a.Type {
		case action.ActionMove:
			dest := shortenDest(a.Dest)
			fmt.Printf("    %s %s %s %s\n", dimStyle.Render(fmt.Sprintf("%2d.", idx)), pathStyle.Render(name), dimStyle.Render("->"), dest)
		case action.ActionDelete:
			fmt.Printf("    %s %s  %s\n", dimStyle.Render(fmt.Sprintf("%2d.", idx)), pathStyle.Render(name), dimStyle.Render(a.Reason))
		case action.ActionRename:
			dest := filepath.Base(a.Dest)
			fmt.Printf("    %s %s %s %s\n", dimStyle.Render(fmt.Sprintf("%2d.", idx)), pathStyle.Render(name), dimStyle.Render("->"), dest)
		}
		idx++
	}
}

func shortenDest(dest string) string {
	// Show last 2-3 path segments for readability
	parts := strings.Split(filepath.ToSlash(dest), "/")
	if len(parts) > 3 {
		return filepath.Join(parts[len(parts)-3:]...)
	}
	return dest
}

// Confirm launches the interactive TUI for approving actions.
func Confirm(actions []action.Action) []action.Action {
	return ConfirmTUI(actions)
}

func countActions(actions []action.Action) (moves, deletes, renames int) {
	for _, a := range actions {
		switch a.Type {
		case action.ActionMove:
			moves++
		case action.ActionDelete:
			deletes++
		case action.ActionRename:
			renames++
		}
	}
	return
}
