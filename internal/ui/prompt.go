package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wizli595/tidydir/internal/action"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	moveStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	deleteStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	renameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	pathStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	countStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	promptStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

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

func Confirm(actions []action.Action) []action.Action {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("  %s\n", promptStyle.Render("Approve?"))
	fmt.Printf("    %s approve all\n", dimStyle.Render("[a]"))
	fmt.Printf("    %s reject all\n", dimStyle.Render("[n]"))
	fmt.Printf("    %s pick one by one\n\n", dimStyle.Render("[i]"))
	fmt.Print("  > ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "a", "all", "y", "yes":
		return actions
	case "n", "none", "no":
		return nil
	case "i", "interactive":
		return confirmInteractive(actions, reader)
	default:
		fmt.Println("  Aborted.")
		return nil
	}
}

func confirmInteractive(actions []action.Action, reader *bufio.Reader) []action.Action {
	var approved []action.Action
	total := len(actions)

	fmt.Println()
	for i, a := range actions {
		progress := dimStyle.Render(fmt.Sprintf("[%d/%d]", i+1, total))
		name := filepath.Base(a.Source)

		switch a.Type {
		case action.ActionMove:
			dest := shortenDest(a.Dest)
			fmt.Printf("  %s %s %s %s %s\n", progress, moveStyle.Render("MOVE"), pathStyle.Render(name), dimStyle.Render("->"), dest)
		case action.ActionDelete:
			fmt.Printf("  %s %s %s  %s\n", progress, deleteStyle.Render("DEL "), pathStyle.Render(name), dimStyle.Render(a.Reason))
		case action.ActionRename:
			dest := filepath.Base(a.Dest)
			fmt.Printf("  %s %s %s %s %s\n", progress, renameStyle.Render("REN "), pathStyle.Render(name), dimStyle.Render("->"), dest)
		}

		fmt.Print("         [y]es  [n]o  [q]uit > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			approved = append(approved, a)
		case "q", "quit":
			fmt.Printf("\n  Stopped. %d approved so far.\n", len(approved))
			return approved
		}
	}

	fmt.Printf("\n  %s approved out of %d.\n", countStyle.Render(fmt.Sprintf("%d", len(approved))), total)
	return approved
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
