package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/executor"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
	"github.com/wizli595/tidydir/internal/ui"
)

var organizeCmd = &cobra.Command{
	Use:   "organize [path]",
	Short: "Scan, plan, confirm, and organize a directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		// Scan
		entries, err := scanner.Scan(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d entries in %s\n\n", len(entries), path)

		// Classify
		classifiers := classifier.DefaultClassifiers()
		classifications := classifier.RunAll(classifiers, entries)

		// Plan
		actions := planner.Plan(classifications, path)

		if len(actions) == 0 {
			fmt.Println("Everything looks tidy! Nothing to do.")
			return
		}

		// Show & confirm
		ui.ShowPlan(actions)
		approved := ui.Confirm(actions)

		if len(approved) == 0 {
			fmt.Println("No actions approved. Exiting.")
			return
		}

		// Execute
		if err := executor.Run(approved, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDone! %d actions executed. Undo log saved.\n", len(approved))
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)
}
