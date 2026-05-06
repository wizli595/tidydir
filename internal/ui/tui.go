package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wizli595/tidydir/internal/action"
)

type model struct {
	actions  []action.Action
	selected []bool
	cursor   int
	done     bool
	aborted  bool
}

func newModel(actions []action.Action) model {
	selected := make([]bool, len(actions))
	for i := range selected {
		selected[i] = true
	}
	return model{actions: actions, selected: selected}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "a":
			for i := range m.selected {
				m.selected[i] = true
			}
		case "n":
			for i := range m.selected {
				m.selected[i] = false
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(promptStyle.Render("  Approve actions"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  space: toggle | a: all | n: none | enter: confirm | q: quit"))
	b.WriteString("\n\n")

	// Viewport: show max 15 items around cursor
	maxVisible := 15
	start, end := 0, len(m.actions)
	if len(m.actions) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.actions) {
			end = len(m.actions)
			start = end - maxVisible
		}
	}

	if start > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("    ... %d more above", start)))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		a := m.actions[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		check := "[ ]"
		checkStyle := dimStyle
		if m.selected[i] {
			check = "[x]"
			checkStyle = promptStyle
		}

		name := filepath.Base(a.Source)
		switch a.Type {
		case action.ActionMove:
			dest := shortenDest(a.Dest)
			fmt.Fprintf(&b, "  %s %s  %s %s %s %s\n",
				dimStyle.Render(cursor), checkStyle.Render(check),
				moveStyle.Render("MOVE"), pathStyle.Render(name),
				dimStyle.Render("->"), dest)
		case action.ActionDelete:
			fmt.Fprintf(&b, "  %s %s  %s %s  %s\n",
				dimStyle.Render(cursor), checkStyle.Render(check),
				deleteStyle.Render("DEL "), pathStyle.Render(name),
				dimStyle.Render(a.Reason))
		case action.ActionRename:
			dest := filepath.Base(a.Dest)
			fmt.Fprintf(&b, "  %s %s  %s %s %s %s\n",
				dimStyle.Render(cursor), checkStyle.Render(check),
				renameStyle.Render("REN "), pathStyle.Render(name),
				dimStyle.Render("->"), dest)
		}
	}

	if end < len(m.actions) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("    ... %d more below", len(m.actions)-end)))
		b.WriteString("\n")
	}

	count := 0
	for _, s := range m.selected {
		if s {
			count++
		}
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s selected\n", countStyle.Render(fmt.Sprintf("%d/%d", count, len(m.actions))))

	return b.String()
}

// ConfirmTUI runs an interactive TUI for approving actions.
func ConfirmTUI(actions []action.Action) []action.Action {
	m := newModel(actions)
	p := tea.NewProgram(m)

	result, err := p.Run()
	if err != nil {
		fmt.Printf("  TUI error: %v\n", err)
		return nil
	}

	final := result.(model)
	if final.aborted {
		fmt.Println("  Aborted.")
		return nil
	}

	var approved []action.Action
	for i, a := range final.actions {
		if final.selected[i] {
			approved = append(approved, a)
		}
	}

	fmt.Printf("\n  %s approved out of %d.\n",
		countStyle.Render(fmt.Sprintf("%d", len(approved))), len(final.actions))
	return approved
}
