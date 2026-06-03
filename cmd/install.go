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
	Long: `Installs replay AI skill files into detected AI tool directories.

The skill files teach AI coding agents (Claude Code, Cursor, OpenCode, Windsurf, etc.)
how to generate replay workflow YAML files from plain-language descriptions.

Detected tools:
  - Claude Code (~/.claude/)
  - OpenCode (~/.config/opencode/)
  - Cursor (.cursor/)
  - Windsurf (.windsurf/)

To use after installation, ask your AI agent:
  "Generate a replay workflow to test the login endpoint"`,
	Args:          cobra.ExactArgs(1),
	ValidArgs:     []string{"ai-skills"},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		all := ai.ListAllTools()

		fmt.Fprintf(cmd.OutOrStdout(), "Scanning for AI tools...\n\n")

		var detectedNames []string
		for _, t := range all {
			status := "✗"
			if t.Exists {
				status = "✓"
				detectedNames = append(detectedNames, t.Name)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", status, t.Name)
		}

		fmt.Fprintln(cmd.OutOrStdout())

		if len(detectedNames) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No supported AI tools detected.\n")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Install to all detected tools? [Y/n] ")
		reader := bufio.NewReader(cmd.InOrStdin())
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "n" || answer == "no" {
			fmt.Fprintf(cmd.OutOrStdout(), "Skipped.\n")
			return nil
		}

		targets := ai.DetectTargets()
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

func init() {
	rootCmd.AddCommand(installCmd)
}
