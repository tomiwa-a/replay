package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillFS_ContainsExpectedFiles(t *testing.T) {
	entries, err := SkillFS.ReadDir("skill")
	if err != nil {
		t.Fatalf("failed to read skill directory: %v", err)
	}

	expected := map[string]bool{
		"SKILL.md":              false,
		"Workflow-Reference.md": false,
		"Test-Patterns.md":      false,
		"Template-Reference.md": false,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			expected[entry.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected skill file %s not found in embedded FS", name)
		}
	}
}

func TestInstall_CreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	targets := []InstallTarget{
		{
			Name: "TestTool",
			Dir:  filepath.Join(tmpDir, "test-tool", "skills", "replay"),
		},
	}

	installed, err := Install(targets)
	if err != nil {
		t.Fatalf("Install() returned error: %v", err)
	}

	expectedFiles := []string{
		"SKILL.md",
		"Workflow-Reference.md",
		"Test-Patterns.md",
		"Template-Reference.md",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, "test-tool", "skills", "replay", f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", path)
		}
	}

	if len(installed) != len(expectedFiles) {
		t.Errorf("expected %d installed files, got %d", len(expectedFiles), len(installed))
	}
}

func TestInstall_MultipleTargets(t *testing.T) {
	tmpDir := t.TempDir()
	targets := []InstallTarget{
		{Name: "Tool1", Dir: filepath.Join(tmpDir, "tool1")},
		{Name: "Tool2", Dir: filepath.Join(tmpDir, "tool2")},
	}

	installed, err := Install(targets)
	if err != nil {
		t.Fatalf("Install() returned error: %v", err)
	}

	if len(installed) != 8 {
		t.Errorf("expected 8 installed files, got %d", len(installed))
	}

	for _, path := range installed {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("installed file %s does not exist", path)
		}
	}
}

func TestInstall_FileContent(t *testing.T) {
	tmpDir := t.TempDir()
	targets := []InstallTarget{
		{Name: "Test", Dir: filepath.Join(tmpDir, "target")},
	}

	installed, err := Install(targets)
	if err != nil {
		t.Fatalf("Install() returned error: %v", err)
	}

	for _, path := range installed {
		if filepath.Base(path) != "SKILL.md" {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read installed file: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "name: replay") {
			t.Errorf("installed SKILL.md missing frontmatter 'name: replay'")
		}
		if !strings.Contains(content, "Replay Testing Skill") {
			t.Errorf("installed SKILL.md missing title 'Replay Testing Skill'")
		}
		if !strings.Contains(content, "Workflow-Reference.md") {
			t.Errorf("installed SKILL.md missing reference to Workflow-Reference.md")
		}
		if !strings.Contains(content, "docs.ellomas.com/replay") {
			t.Errorf("installed SKILL.md missing docs URL reference")
		}
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()

	if dirExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("dirExists should return false for nonexistent directory")
	}

	if !dirExists(tmpDir) {
		t.Error("dirExists should return true for existing directory")
	}

	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if dirExists(filePath) {
		t.Error("dirExists should return false for a file")
	}
}
