package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/executor"
)

var undoCmd = &cobra.Command{
	Use:   "undo [path]",
	Short: "Revert the last organize operation",
	Long: `Reverses all actions from the last organize run.
Restores moved files, recovers trashed items, and cleans up
empty directories created during organize.

Examples:
  tidydir undo ~/Documents`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if err := executor.Undo(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error undoing: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Undo complete. Files restored to original locations.")
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
