package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
	"github.com/wizli595/tidydir/internal/ui"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a directory and show the organization plan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		// Step 1: Scan
		entries, err := scanner.Scan(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d entries in %s\n\n", len(entries), path)

		// Step 2: Classify
		classifiers := classifier.DefaultClassifiers()
		classifications := classifier.RunAll(classifiers, entries)

		// Step 3: Plan
		actions := planner.Plan(classifications, path)

		if len(actions) == 0 {
			fmt.Println("Everything looks tidy! Nothing to do.")
			return
		}

		// Step 4: Show plan
		ui.ShowPlan(actions)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
