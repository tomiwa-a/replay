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

// ToolDef describes a supported AI tool and how to detect it.
type ToolDef struct {
	Name    string // Human-readable tool name
	Dir     string // Destination directory for skill files
	Exists  bool   // Whether the tool's config directory was found
	Check   string // The path checked for existence
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ListAllTools returns all known AI tools with their detection status.
func ListAllTools() []ToolDef {
	home, _ := os.UserHomeDir()

	tools := []ToolDef{
		{
			Name:   "Claude Code",
			Dir:    filepath.Join(home, ".claude", "skills", "replay"),
			Check:  filepath.Join(home, ".claude"),
		},
		{
			Name:   "OpenCode",
			Dir:    filepath.Join(home, ".config", "opencode", "skill", "replay"),
			Check:  filepath.Join(home, ".config", "opencode"),
		},
		{
			Name:   "Gemini CLI",
			Dir:    filepath.Join(home, ".gemini", "skills", "replay"),
			Check:  filepath.Join(home, ".gemini"),
		},
		{
			Name:   "Cursor",
			Dir:    filepath.Join(".cursor", "rules", "replay"),
			Check:  ".cursor",
		},
		{
			Name:   "Windsurf",
			Dir:    filepath.Join(".windsurf", "rules", "replay"),
			Check:  ".windsurf",
		},
		{
			Name:   "Antigravity",
			Dir:    filepath.Join(".agent", "skills", "replay"),
			Check:  ".agent",
		},
	}

	for i := range tools {
		tools[i].Exists = dirExists(tools[i].Check)
	}

	return tools
}

// DetectTargets returns only the detected (installed) tools.
func DetectTargets() []InstallTarget {
	all := ListAllTools()
	var targets []InstallTarget
	for _, t := range all {
		if t.Exists {
			targets = append(targets, InstallTarget{Name: t.Name, Dir: t.Dir})
		}
	}
	return targets
}

// LookupTargets returns InstallTargets for the given tool names (case-insensitive).
// Unknown names are ignored.
func LookupTargets(names []string) []InstallTarget {
	all := ListAllTools()
	normalized := make(map[string]ToolDef)
	for _, t := range all {
		normalized[strings.ToLower(t.Name)] = t
	}

	var targets []InstallTarget
	for _, name := range names {
		if t, ok := normalized[strings.ToLower(strings.TrimSpace(name))]; ok {
			targets = append(targets, InstallTarget{Name: t.Name, Dir: t.Dir})
		}
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
