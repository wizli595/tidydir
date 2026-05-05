package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tidydir",
	Short: "Smart directory organizer — scan, plan, confirm, execute",
	Long:  "tidydir scans messy directories, classifies files/projects, shows you a plan, and organizes on your confirmation.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
