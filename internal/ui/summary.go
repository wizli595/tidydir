package ui

import (
	"fmt"
	"strings"

	"github.com/wizli595/tidydir/internal/insights"
)

// ShowSummary displays execution results.
func ShowSummary(moved, deleted, renamed int, freed int64) {
	fmt.Println()
	fmt.Println(titleStyle.Render("  SUMMARY"))
	fmt.Println(dimStyle.Render("  " + strings.Repeat("-", 56)))

	if moved > 0 {
		fmt.Printf("  %s %s\n", moveStyle.Render(fmt.Sprintf("%d", moved)), dimStyle.Render("files moved"))
	}
	if deleted > 0 {
		fmt.Printf("  %s %s\n", deleteStyle.Render(fmt.Sprintf("%d", deleted)), dimStyle.Render("files trashed"))
	}
	if renamed > 0 {
		fmt.Printf("  %s %s\n", renameStyle.Render(fmt.Sprintf("%d", renamed)), dimStyle.Render("files renamed"))
	}
	if freed > 0 {
		fmt.Printf("  %s %s\n", countStyle.Render(insights.FormatSize(freed)), dimStyle.Render("freed"))
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("  Undo log saved. Run 'tidydir undo <path>' to reverse."))
}
