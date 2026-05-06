package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate completion scripts for your shell.

Usage:
  # Bash
  tidydir completion bash > /etc/bash_completion.d/tidydir

  # Zsh
  tidydir completion zsh > "${fpath[1]}/_tidydir"

  # Fish
  tidydir completion fish > ~/.config/fish/completions/tidydir.fish

  # PowerShell
  tidydir completion powershell | Out-String | Invoke-Expression`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			rootCmd.GenPowerShellCompletion(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
