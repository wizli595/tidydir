package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Monitor a directory and report when it needs organizing",
	Long: `Watches a directory for changes and reports when new files
need organizing. Press Ctrl+C to stop.

Examples:
  tidydir watch ~/Documents
  tidydir watch ~/Downloads`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		cfg := config.Load(path)

		reportStatus(path, cfg)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer watcher.Close()

		if err := watcher.Add(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error watching %s: %v\n", path, err)
			os.Exit(1)
		}

		fmt.Printf("Watching %s for changes... (Ctrl+C to stop)\n\n", path)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		debounce := time.NewTimer(0)
		debounce.Stop()

		for {
			select {
			case <-stop:
				fmt.Println("\nStopped watching.")
				return
			case event := <-watcher.Events:
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					debounce.Reset(2 * time.Second)
				}
			case err := <-watcher.Errors:
				fmt.Fprintf(os.Stderr, "Watch error: %v\n", err)
			case <-debounce.C:
				reportStatus(path, cfg)
			}
		}
	},
}

func reportStatus(path string, cfg *config.Config) {
	entries, err := scanner.Scan(path, scanner.Options{Ignore: cfg.Ignore})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
		return
	}

	classifiers := classifier.DefaultClassifiers(cfg.ProjectMarkers, cfg.CustomClassifiers)
	classifications := classifier.RunAll(classifiers, entries)
	actions := planner.Plan(classifications, path, cfg.Folders, cfg.CustomRules)
	renames := planner.PlanRenames(entries, path, actions)
	actions = append(actions, renames...)

	now := time.Now().Format("15:04:05")

	if len(actions) == 0 {
		fmt.Printf("[%s] %d entries -- all tidy.\n", now, len(entries))
		return
	}

	moves, deletes, renameCount := 0, 0, 0
	for _, a := range actions {
		switch a.Type {
		case action.ActionMove:
			moves++
		case action.ActionDelete:
			deletes++
		case action.ActionRename:
			renameCount++
		}
	}

	parts := []string{}
	if moves > 0 {
		parts = append(parts, fmt.Sprintf("%d moves", moves))
	}
	if deletes > 0 {
		parts = append(parts, fmt.Sprintf("%d deletes", deletes))
	}
	if renameCount > 0 {
		parts = append(parts, fmt.Sprintf("%d renames", renameCount))
	}

	fmt.Printf("[%s] %d entries, %d need organizing (%s)\n",
		now, len(entries), len(actions), strings.Join(parts, ", "))
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
