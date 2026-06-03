package cmd

import (
	"fmt"

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
  - Windsurf (.windsurf/)`,
	Args:          cobra.ExactArgs(1),
	ValidArgs:     []string{"ai-skills"},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		targets := ai.DetectTargets()

		if len(targets) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No supported AI tools detected.\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Install Claude Code, Cursor, or OpenCode, then re-run this command.\n")
			fmt.Fprintf(cmd.OutOrStdout(), "You can also copy the skill files manually from:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  https://github.com/tomiwa-a/replay/tree/main/internal/ai/skill\n")
			return nil
		}

		installed, err := ai.Install(targets)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✓ Installed replay skill for %d tool(s):\n", len(targets))
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
