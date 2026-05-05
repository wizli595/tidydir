package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/executor"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last organize operation",
	Run: func(cmd *cobra.Command, args []string) {
		if err := executor.Undo(); err != nil {
			fmt.Fprintf(os.Stderr, "Error undoing: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Undo complete. Files restored to original locations.")
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
