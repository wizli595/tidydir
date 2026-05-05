package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/executor"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
	"github.com/wizli595/tidydir/internal/ui"
)

var organizeCmd = &cobra.Command{
	Use:   "organize [path]",
	Short: "Scan, plan, confirm, and organize a directory",
	Long:  "Scans the target directory, classifies entries, shows a plan, asks for confirmation, then executes the approved actions.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		depth, _ := cmd.Flags().GetInt("depth")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cfg := config.Load(path)

		entries, err := scanner.Scan(path, scanner.Options{
			Ignore: cfg.Ignore,
			Depth:  depth,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d entries in %s\n\n", len(entries), path)

		classifiers := classifier.DefaultClassifiers(cfg.ProjectMarkers)
		classifications := classifier.RunAll(classifiers, entries)
		actions := planner.Plan(classifications, path, cfg.Folders, cfg.CustomRules)

		if len(actions) == 0 {
			fmt.Println("Everything looks tidy! Nothing to do.")
			return
		}

		ui.ShowPlan(actions)

		if dryRun {
			return
		}

		approved := ui.Confirm(actions)
		if len(approved) == 0 {
			fmt.Println("No actions approved. Exiting.")
			return
		}

		if err := executor.Run(approved, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDone! %d actions executed. Undo log saved.\n", len(approved))
	},
}

func init() {
	organizeCmd.Flags().Int("depth", 0, "Recursion depth (0 = top-level only)")
	organizeCmd.Flags().Bool("dry-run", false, "Show plan without executing")
	rootCmd.AddCommand(organizeCmd)
}
