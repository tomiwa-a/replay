package ai

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed skill/*.md
var SkillFS embed.FS

// InstallTarget represents a detected AI tool and its skill directory.
type InstallTarget struct {
	Name string // Human-readable tool name
	Dir  string // Destination directory for skill files
}

// DetectTargets checks the current environment for supported AI tools
// and returns the directories where replay skills should be installed.
func DetectTargets() []InstallTarget {
	var targets []InstallTarget

	home, _ := os.UserHomeDir()

	if claudeDir := filepath.Join(home, ".claude"); dirExists(claudeDir) {
		targets = append(targets, InstallTarget{
			Name: "Claude Code",
			Dir:  filepath.Join(claudeDir, "skills", "replay"),
		})
	}

	if opencodeDir := filepath.Join(home, ".config", "opencode"); dirExists(opencodeDir) {
		targets = append(targets, InstallTarget{
			Name: "OpenCode",
			Dir:  filepath.Join(opencodeDir, "skill", "replay"),
		})
	}

	if cursorDir := filepath.Join(".cursor"); dirExists(cursorDir) {
		targets = append(targets, InstallTarget{
			Name: "Cursor",
			Dir:  filepath.Join(cursorDir, "rules", "replay"),
		})
	}

	if windsurfDir := filepath.Join(".windsurf"); dirExists(windsurfDir) {
		targets = append(targets, InstallTarget{
			Name: "Windsurf",
			Dir:  filepath.Join(windsurfDir, "rules", "replay"),
		})
	}

	return targets
}

// Install copies the embedded skill files to each target directory.
// Returns a list of installed file paths and any errors encountered.
func Install(targets []InstallTarget) ([]string, error) {
	var installed []string

	skillEntries, err := fs.Glob(SkillFS, "skill/*.md")
	if err != nil {
		return nil, fmt.Errorf("list skill files: %w", err)
	}

	for _, target := range targets {
		for _, entry := range skillEntries {
			data, err := SkillFS.ReadFile(entry)
			if err != nil {
				return installed, fmt.Errorf("read %s: %w", entry, err)
			}

			filename := strings.TrimPrefix(entry, "skill/")

			if err := os.MkdirAll(target.Dir, 0755); err != nil {
				return installed, fmt.Errorf("create dir %s: %w", target.Dir, err)
			}

			destPath := filepath.Join(target.Dir, filename)
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return installed, fmt.Errorf("write %s: %w", destPath, err)
			}
			installed = append(installed, destPath)
		}
	}

	return installed, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
