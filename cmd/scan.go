package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
	"github.com/wizli595/tidydir/internal/ui"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a directory and show the organization plan (dry-run)",
	Long:  "Scans the target directory, classifies all entries, and displays the proposed actions without making any changes.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		depth, _ := cmd.Flags().GetInt("depth")

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
	},
}

func init() {
	scanCmd.Flags().Int("depth", 0, "Recursion depth (0 = top-level only)")
	rootCmd.AddCommand(scanCmd)
}
