package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/replay/replay/internal/workflow"
)

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "replay.yaml")
	content := `version: v1
config:
  http:
    base_url: https://api.example.com
  vars:
    env: staging
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Version != "v1" {
		t.Errorf("expected v1, got %s", cfg.Version)
	}
	if cfg.Config.HTTP.BaseURL != "https://api.example.com" {
		t.Errorf("expected base_url, got %s", cfg.Config.HTTP.BaseURL)
	}
	if cfg.Config.Vars["env"] != "staging" {
		t.Errorf("expected env=staging, got %v", cfg.Config.Vars["env"])
	}
}

func TestLoadConfigFileWithProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "replay.yaml")
	content := `version: v1
config:
  vars:
    env: default
profiles:
  dev:
    config:
      http:
        base_url: http://localhost:3000
      vars:
        env: development
  prod:
    config:
      http:
        base_url: https://api.example.com
      vars:
        env: production
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	profile, err := cfg.GetProfile("dev")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Config.HTTP.BaseURL != "http://localhost:3000" {
		t.Errorf("expected dev base_url, got %s", profile.Config.HTTP.BaseURL)
	}
}

func TestLoadConfigFileWithPresets(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "replay.yaml")
	content := `version: v1
presets:
  auth:
    config:
      vars:
        auth_url: https://auth.example.com
        client_id: my-client
  database:
    config:
      postgres:
        dsn: postgres://localhost:5432/testdb?sslmode=disable
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	preset, err := cfg.GetPreset("auth")
	if err != nil {
		t.Fatal(err)
	}
	if preset.Config.Vars["auth_url"] != "https://auth.example.com" {
		t.Errorf("expected auth_url, got %v", preset.Config.Vars["auth_url"])
	}
}

func TestMergeConfigs(t *testing.T) {
	base := workflow.Config{
		HTTP: workflow.HTTPConfig{
			BaseURL: "https://base.example.com",
			Headers: map[string]string{
				"X-Base": "true",
			},
		},
		Vars: map[string]any{
			"a": 1,
			"b": 2,
		},
	}

	override := workflow.Config{
		HTTP: workflow.HTTPConfig{
			BaseURL: "https://override.example.com",
			Headers: map[string]string{
				"X-Override": "true",
			},
		},
		Vars: map[string]any{
			"b": 3,
			"c": 4,
		},
	}

	merged := MergeConfigs(base, override)

	if merged.HTTP.BaseURL != "https://override.example.com" {
		t.Errorf("expected override base_url, got %s", merged.HTTP.BaseURL)
	}
	if merged.HTTP.Headers["X-Base"] != "true" {
		t.Error("base headers should be preserved")
	}
	if merged.HTTP.Headers["X-Override"] != "true" {
		t.Error("override headers should be added")
	}
	if merged.Vars["a"] != 1 {
		t.Error("base var a should be preserved")
	}
	if merged.Vars["b"] != 3 {
		t.Error("override var b should win")
	}
	if merged.Vars["c"] != 4 {
		t.Error("override var c should be added")
	}
}

func TestMergeConfigsPostgres(t *testing.T) {
	base := workflow.Config{
		Postgres: workflow.PostgresConfig{DSN: "postgres://base"},
	}
	override := workflow.Config{
		Postgres: workflow.PostgresConfig{DSN: "postgres://override"},
	}

	merged := MergeConfigs(base, override)
	if merged.Postgres.DSN != "postgres://override" {
		t.Errorf("expected override DSN, got %s", merged.Postgres.DSN)
	}
}

func TestMergeConfigsRedis(t *testing.T) {
	base := workflow.Config{
		Redis: workflow.RedisConfig{Addr: "redis://base"},
	}
	override := workflow.Config{
		Redis: workflow.RedisConfig{Addr: "redis://override"},
	}

	merged := MergeConfigs(base, override)
	if merged.Redis.Addr != "redis://override" {
		t.Errorf("expected override addr, got %s", merged.Redis.Addr)
	}
}

func TestMergeVars(t *testing.T) {
	base := map[string]any{"a": 1, "b": 2}
	override := map[string]any{"b": 3, "c": 4}

	merged := mergeVars(base, override)
	if merged["a"] != 1 {
		t.Error("base var a should be preserved")
	}
	if merged["b"] != 3 {
		t.Error("override var b should win")
	}
	if merged["c"] != 4 {
		t.Error("override var c should be added")
	}
}

func TestMergeVarsNil(t *testing.T) {
	merged := mergeVars(nil, nil)
	if merged != nil {
		t.Error("expected nil")
	}
}

func TestMergeValidate(t *testing.T) {
	base := []workflow.VarDef{
		{Name: "a", Type: "string"},
		{Name: "b", Type: "number"},
	}
	override := []workflow.VarDef{
		{Name: "b", Type: "string"},
		{Name: "c", Type: "bool"},
	}

	merged := mergeValidate(base, override)
	if len(merged) != 3 {
		t.Fatalf("expected 3, got %d", len(merged))
	}

	found := make(map[string]bool)
	for _, v := range merged {
		found[v.Name] = true
	}
	if !found["a"] || !found["b"] || !found["c"] {
		t.Error("expected a, b, c")
	}
}

func TestApplyConfigToFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "replay.yaml")
	content := `version: v1
config:
  http:
    base_url: https://global.example.com
  vars:
    global_var: from_config
profiles:
  dev:
    config:
      http:
        base_url: http://localhost:3000
      vars:
        env: dev
presets:
  auth:
    config:
      vars:
        auth_url: https://auth.example.com
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	wf := &workflow.Workflow{
		Name: "test",
		Config: workflow.Config{
			Vars: map[string]any{
				"wf_var": "from_workflow",
			},
		},
	}

	if err := ApplyConfigToFile(cfg, wf, "dev"); err != nil {
		t.Fatal(err)
	}

	if wf.Config.HTTP.BaseURL != "http://localhost:3000" {
		t.Errorf("expected dev base_url, got %s", wf.Config.HTTP.BaseURL)
	}
	if wf.Config.Vars["global_var"] != "from_config" {
		t.Error("global var should be preserved")
	}
	if wf.Config.Vars["env"] != "dev" {
		t.Error("profile var should be set")
	}
	if wf.Config.Vars["wf_var"] != "from_workflow" {
		t.Error("workflow var should be preserved")
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	configPath := filepath.Join(tmpDir, "replay.yaml")
	if err := os.WriteFile(configPath, []byte("version: v1"), 0644); err != nil {
		t.Fatal(err)
	}

	found := FindConfigFile()
	if found != "replay.yaml" {
		t.Errorf("expected replay.yaml, got %s", found)
	}
}

func TestFindConfigFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	found := FindConfigFile()
	if found != "" {
		t.Errorf("expected empty, got %s", found)
	}
}

func TestGetProfileNotFound(t *testing.T) {
	cfg := &ConfigFile{
		Profiles: map[string]Profile{},
	}
	_, err := cfg.GetProfile("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestGetPresetNotFound(t *testing.T) {
	cfg := &ConfigFile{
		Presets: map[string]Preset{},
	}
	_, err := cfg.GetPreset("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent preset")
	}
}
