package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wizli595/tidydir/internal/action"
)

var (
	moveStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	deleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	renameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
)

// ShowPlan prints the proposed actions in a readable format.
func ShowPlan(actions []action.Action) {
	fmt.Printf("📋 Plan: %d actions\n", len(actions))
	fmt.Println(strings.Repeat("─", 60))

	for i, a := range actions {
		prefix := fmt.Sprintf("[%d]", i+1)
		switch a.Type {
		case action.ActionMove:
			fmt.Printf("%s %s %s\n", prefix, moveStyle.Render("MOVE"), a.Source)
			fmt.Printf("     → %s\n", a.Dest)
		case action.ActionDelete:
			fmt.Printf("%s %s %s\n", prefix, deleteStyle.Render("DELETE"), a.Source)
		case action.ActionRename:
			fmt.Printf("%s %s %s\n", prefix, renameStyle.Render("RENAME"), a.Source)
			fmt.Printf("     → %s\n", a.Dest)
		}
		fmt.Printf("     %s\n", dimStyle.Render(a.Reason))
	}

	fmt.Println(strings.Repeat("─", 60))
}

// Confirm asks the user to approve actions.
// Returns only the approved actions.
func Confirm(actions []action.Action) []action.Action {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nApprove actions? [a]ll / [n]one / [i]nteractive")
	fmt.Print("> ")
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
		fmt.Println("Invalid choice, aborting.")
		return nil
	}
}

func confirmInteractive(actions []action.Action, reader *bufio.Reader) []action.Action {
	var approved []action.Action

	for i, a := range actions {
		fmt.Printf("\n[%d/%d] ", i+1, len(actions))
		switch a.Type {
		case action.ActionMove:
			fmt.Printf("%s %s → %s\n", moveStyle.Render("MOVE"), a.Source, a.Dest)
		case action.ActionDelete:
			fmt.Printf("%s %s\n", deleteStyle.Render("DELETE"), a.Source)
		case action.ActionRename:
			fmt.Printf("%s %s → %s\n", renameStyle.Render("RENAME"), a.Source, a.Dest)
		}

		fmt.Print("  approve? [y/n/q]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			approved = append(approved, a)
		case "q", "quit":
			return approved
		}
	}

	return approved
}
