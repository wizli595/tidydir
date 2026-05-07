package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wizli595/tidydir/internal/action"
	"github.com/wizli595/tidydir/internal/classifier"
	"github.com/wizli595/tidydir/internal/config"
	"github.com/wizli595/tidydir/internal/planner"
	"github.com/wizli595/tidydir/internal/scanner"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule [path]",
	Short: "Set up scheduled auto-scan with desktop notifications",
	Long: `Generates a system scheduler command to run tidydir periodically,
or runs a single check with a desktop notification.

Examples:
  tidydir schedule ~/Documents               Show setup instructions
  tidydir schedule ~/Documents --run          Run check + notify now
  tidydir schedule ~/Documents --every 1h     Custom interval`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		runNow, _ := cmd.Flags().GetBool("run")
		every, _ := cmd.Flags().GetString("every")

		if runNow {
			runScheduledCheck(path)
			return
		}

		showSetupInstructions(path, every)
	},
}

func runScheduledCheck(path string) {
	cfg := config.Load(path)
	entries, err := scanner.Scan(path, scanner.Options{Ignore: cfg.Ignore})
	if err != nil {
		return
	}

	classifiers := classifier.DefaultClassifiers(cfg.ProjectMarkers, cfg.CustomClassifiers)
	classifications := classifier.RunAll(classifiers, entries)
	actions := planner.Plan(classifications, path, cfg.Folders, cfg.CustomRules)
	renames := planner.PlanRenames(entries, path, actions)
	actions = append(actions, renames...)

	if len(actions) == 0 {
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
		parts = append(parts, fmt.Sprintf("%d to move", moves))
	}
	if deletes > 0 {
		parts = append(parts, fmt.Sprintf("%d to delete", deletes))
	}
	if renameCount > 0 {
		parts = append(parts, fmt.Sprintf("%d to rename", renameCount))
	}

	msg := fmt.Sprintf("%s needs organizing: %s", path, strings.Join(parts, ", "))
	sendNotification("tidydir", msg)
	fmt.Println(msg)
}

func sendNotification(title, message string) {
	switch runtime.GOOS {
	case "windows":
		script := fmt.Sprintf(
			`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; `+
				`$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent(0); `+
				`$text = $template.GetElementsByTagName('text'); `+
				`$text.Item(0).AppendChild($template.CreateTextNode('%s')) > $null; `+
				`$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('tidydir'); `+
				`$notifier.Show([Windows.UI.Notifications.ToastNotification]::new($template))`,
			strings.ReplaceAll(message, "'", "''"),
		)
		exec.Command("powershell", "-NoProfile", "-Command", script).Run()
	case "darwin":
		script := fmt.Sprintf(
			`display notification "%s" with title "%s"`,
			strings.ReplaceAll(message, `"`, `\"`),
			title,
		)
		exec.Command("osascript", "-e", script).Run()
	default:
		exec.Command("notify-send", title, message).Run()
	}
}

func showSetupInstructions(path, every string) {
	exe, _ := os.Executable()
	if exe == "" {
		exe = "tidydir"
	}

	scheduleCmd := fmt.Sprintf("%s schedule %s --run", exe, path)

	fmt.Println()
	fmt.Println("  SCHEDULED AUTO-SCAN SETUP")
	fmt.Println("  " + strings.Repeat("-", 56))

	switch runtime.GOOS {
	case "windows":
		interval := "60"
		if every != "" {
			interval = parseDurationMinutes(every)
		}
		fmt.Println()
		fmt.Println("  Run this in PowerShell (admin):")
		fmt.Println()
		fmt.Printf("  schtasks /create /tn \"tidydir-scan\" /tr \"%s\" /sc minute /mo %s\n", scheduleCmd, interval)
		fmt.Println()
		fmt.Println("  To remove:")
		fmt.Println("  schtasks /delete /tn \"tidydir-scan\" /f")

	case "darwin", "linux":
		cronInterval := "0 * * * *"
		if every != "" {
			cronInterval = parseCronInterval(every)
		}
		fmt.Println()
		fmt.Println("  Add this to your crontab (crontab -e):")
		fmt.Println()
		fmt.Printf("  %s %s\n", cronInterval, scheduleCmd)
		fmt.Println()
		fmt.Println("  To remove: crontab -e and delete the line")
	}

	fmt.Println()
	fmt.Println("  " + strings.Repeat("-", 56))
	fmt.Println("  Test it now: tidydir schedule " + path + " --run")
}

func parseDurationMinutes(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "h") {
		s = strings.TrimSuffix(s, "h")
		return fmt.Sprintf("%s", multiply(s, 60))
	}
	if strings.HasSuffix(s, "m") {
		return strings.TrimSuffix(s, "m")
	}
	return s
}

func parseCronInterval(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "h") {
		h := strings.TrimSuffix(s, "h")
		return fmt.Sprintf("0 */%s * * *", h)
	}
	if strings.HasSuffix(s, "m") {
		m := strings.TrimSuffix(s, "m")
		return fmt.Sprintf("*/%s * * * *", m)
	}
	return "0 * * * *"
}

func multiply(s string, factor int) string {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return fmt.Sprintf("%d", n*factor)
}

func init() {
	scheduleCmd.Flags().Bool("run", false, "Run a single check and notify")
	scheduleCmd.Flags().String("every", "", "Check interval (e.g. 30m, 1h, 2h)")
	rootCmd.AddCommand(scheduleCmd)
}
