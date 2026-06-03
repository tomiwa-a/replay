package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/replay/replay/internal/ai"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:           "install ai-skills",
	Short:         "Install AI agent skills for replay workflow generation",
	Long: `Installs replay AI skill files into selected AI tool directories.

The skill files teach AI coding agents (Claude Code, Cursor, OpenCode,
Gemini CLI, Windsurf, Antigravity, etc.) how to generate replay workflow
YAML files from plain-language descriptions.

To use after installation, ask your AI agent:
  "Generate a replay workflow to test the login endpoint"`,
	Args:          cobra.ExactArgs(1),
	ValidArgs:     []string{"ai-skills"},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		all := ai.ListAllTools()

		fmt.Fprintf(cmd.OutOrStdout(), "Scanning for AI tools...\n\n")

		for i, t := range all {
			status := "✗"
			if t.Exists {
				status = "✓"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  [%d] [%s] %s\n", i+1, status, t.Name)
		}

		fmt.Fprintln(cmd.OutOrStdout())

		var detectedNames []string
		for _, t := range all {
			if t.Exists {
				detectedNames = append(detectedNames, t.Name)
			}
		}

		if len(detectedNames) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No supported AI tools detected.\n")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Install to (comma-separated numbers/names, or 'all'): ")
		reader := bufio.NewReader(cmd.InOrStdin())
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)

		if answer == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Skipped.\n")
			return nil
		}

		var targets []ai.InstallTarget

		if strings.ToLower(answer) == "all" {
			targets = ai.DetectTargets()
		} else {
			parts := strings.Split(answer, ",")
			var names []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				// Check if it's a number
				if i, ok := parseNumber(p); ok && i >= 1 && i <= len(all) {
					names = append(names, all[i-1].Name)
				} else {
					names = append(names, p)
				}
			}
			targets = ai.LookupTargets(names)
		}

		if len(targets) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No matching tools selected.\n")
			return nil
		}

		installed, err := ai.Install(targets)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\n✓ Installed replay skill for %d tool(s):\n", len(targets))
		for _, path := range installed {
			fmt.Fprintf(cmd.OutOrStdout(), "  • %s\n", path)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nTo use, ask your AI agent:\n")
		fmt.Fprintf(cmd.OutOrStdout(), `  "Generate a replay workflow to test the login endpoint"`+"\n")

		return nil
	},
}

func parseNumber(s string) (int, bool) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

func init() {
	rootCmd.AddCommand(installCmd)
}
