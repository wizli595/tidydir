package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/executor"
	"github.com/wizli595/tidydir/internal/insights"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
	"github.com/wizli595/tidydir/internal/ui"
)

var organizeCmd = &cobra.Command{
	Use:   "organize [paths...]",
	Short: "Organize one or more directories with confirmation",
	Long: `Scans one or more directories, classifies entries, shows a plan,
asks for your confirmation, then executes the approved actions.

Deletes are non-destructive (moved to .tidydir_trash/).
An undo log is saved after execution.

Examples:
  tidydir organize ~/Documents
  tidydir organize ~/Documents ~/Downloads
  tidydir organize ~/Downloads --dry-run
  tidydir organize ~/Projects --depth 1
  tidydir organize ~/Downloads --system-trash
  tidydir organize ~/Documents --profile work`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		depth, _ := cmd.Flags().GetInt("depth")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		systemTrash, _ := cmd.Flags().GetBool("system-trash")
		profile, _ := cmd.Flags().GetString("profile")

		for _, path := range args {
			if len(args) > 1 {
				fmt.Printf("\n  --- %s ---\n", path)
			}
			organizeDir(path, depth, dryRun, systemTrash, profile)
		}
	},
}

func organizeDir(path string, depth int, dryRun, systemTrash bool, profile string) {
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
	ui.ShowInsights(ins)

	actions := planner.Plan(classifications, path, cfg.Folders, cfg.CustomRules)
	renames := planner.PlanRenames(entries, path, actions)
	actions = append(actions, renames...)

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

	opts := executor.RunOptions{SystemTrash: systemTrash}
	stats, err := executor.Run(approved, path, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	ui.ShowSummary(stats.Moved, stats.Deleted, stats.Renamed, stats.Freed)
}

func init() {
	organizeCmd.Flags().Int("depth", 0, "Recursion depth (0 = top-level only)")
	organizeCmd.Flags().Bool("dry-run", false, "Show plan without executing")
	organizeCmd.Flags().Bool("system-trash", false, "Use OS trash (Recycle Bin / macOS Trash)")
	organizeCmd.Flags().String("profile", "", "Config profile to use")
	rootCmd.AddCommand(organizeCmd)
}
