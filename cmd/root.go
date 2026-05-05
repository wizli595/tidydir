package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tidydir",
	Short: "Smart directory organizer",
	Long: `tidydir - Smart Directory Organizer

  Scans messy directories, classifies files and projects by type,
  shows you a plan, and organizes only what you approve.

  Workflow:
    1. tidydir scan <path>        Preview the organization plan
    2. tidydir organize <path>    Execute with confirmation
    3. tidydir undo <path>        Revert the last run

  Configuration:
    Place tidydir.yaml in the target directory to customize rules.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
