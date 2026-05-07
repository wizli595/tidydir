package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/export"
	"github.com/wizli595/tidydir/internal/insights"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
	"github.com/wizli595/tidydir/internal/ui"
)

var scanCmd = &cobra.Command{
	Use:   "scan [paths...]",
	Short: "Preview the organization plan without making changes",
	Long: `Scans one or more directories, classifies all entries, and displays
the proposed actions. No files are moved or deleted.

Examples:
  tidydir scan ~/Documents
  tidydir scan ~/Documents ~/Downloads
  tidydir scan ~/Downloads --depth 2
  tidydir scan ~/Documents --format json
  tidydir scan ~/Documents --format csv
  tidydir scan ~/Documents --profile work`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		depth, _ := cmd.Flags().GetInt("depth")
		format, _ := cmd.Flags().GetString("format")
		profile, _ := cmd.Flags().GetString("profile")

		for _, path := range args {
			if len(args) > 1 {
				fmt.Printf("\n  --- %s ---\n", path)
			}
			scanDir(path, depth, format, profile)
		}
	},
}

func scanDir(path string, depth int, format, profile string) {
	cfg := config.LoadWithProfile(path, profile)

	entries, err := scanner.Scan(path, scanner.Options{
		Ignore: cfg.Ignore,
		Depth:  depth,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		return
	}

	fmt.Printf("  Scanned %d entries in %s\n", len(entries), path)

	classifiers := classifier.DefaultClassifiers(cfg.ProjectMarkers, cfg.CustomClassifiers)
	classifications := classifier.RunAll(classifiers, entries)

	ins := insights.Analyze(entries, classifications, insights.Options{
		LargeFileMB:    cfg.LargeFileMB,
		OldProjectDays: cfg.OldProjectDays,
	})

	actions := planner.Plan(classifications, path, cfg.Folders, cfg.CustomRules)
	renames := planner.PlanRenames(entries, path, actions)
	actions = append(actions, renames...)

	if len(actions) == 0 {
		fmt.Println("Everything looks tidy! Nothing to do.")
		return
	}

	switch format {
	case "json":
		export.ToJSON(os.Stdout, actions)
	case "csv":
		export.ToCSV(os.Stdout, actions)
	default:
		ui.ShowInsights(ins)
		ui.ShowPlan(actions)
	}
}

func init() {
	scanCmd.Flags().Int("depth", 0, "Recursion depth (0 = top-level only)")
	scanCmd.Flags().String("format", "text", "Output format: text, json, csv")
	scanCmd.Flags().String("profile", "", "Config profile to use")
	rootCmd.AddCommand(scanCmd)
}
