package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/replay/replay/internal/workflow"
	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Version  string                `yaml:"version"`
	Config   workflow.Config       `yaml:"config,omitempty"`
	Profiles map[string]Profile    `yaml:"profiles,omitempty"`
	Presets  map[string]Preset     `yaml:"presets,omitempty"`
}

type Profile struct {
	Config workflow.Config `yaml:"config,omitempty"`
}

type Preset struct {
	Config workflow.Config `yaml:"config,omitempty"`
}

func LoadFromFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	var cfg ConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}

	return &cfg, nil
}

func FindConfigFile() string {
	candidates := []string{
		"replay.yaml",
		"replay.yml",
		".replay.yaml",
		".replay.yml",
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func FindConfigFileInDir(dir string) string {
	candidates := []string{
		"replay.yaml",
		"replay.yml",
		".replay.yaml",
		".replay.yml",
	}

	for _, c := range candidates {
		path := filepath.Join(dir, c)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (cfg *ConfigFile) GetProfile(name string) (*Profile, error) {
	if cfg.Profiles == nil {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	p, ok := cfg.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return &p, nil
}

func (cfg *ConfigFile) GetPreset(name string) (*Preset, error) {
	if cfg.Presets == nil {
		return nil, fmt.Errorf("preset %q not found", name)
	}
	p, ok := cfg.Presets[name]
	if !ok {
		return nil, fmt.Errorf("preset %q not found", name)
	}
	return &p, nil
}

func MergeConfigs(base, override workflow.Config) workflow.Config {
	result := workflow.Config{}

	result.HTTP = mergeHTTPConfig(base.HTTP, override.HTTP)
	result.Postgres = mergePostgresConfig(base.Postgres, override.Postgres)
	result.Redis = mergeRedisConfig(base.Redis, override.Redis)
	result.Vars = mergeVars(base.Vars, override.Vars)
	result.Validate = mergeValidate(base.Validate, override.Validate)

	return result
}

func mergeHTTPConfig(base, override workflow.HTTPConfig) workflow.HTTPConfig {
	result := base
	if override.BaseURL != "" {
		result.BaseURL = override.BaseURL
	}
	if override.Debug {
		result.Debug = override.Debug
	}
	if override.Headers != nil {
		if result.Headers == nil {
			result.Headers = make(map[string]string)
		}
		for k, v := range override.Headers {
			result.Headers[k] = v
		}
	}
	return result
}

func mergePostgresConfig(base, override workflow.PostgresConfig) workflow.PostgresConfig {
	if override.DSN != "" {
		return override
	}
	return base
}

func mergeRedisConfig(base, override workflow.RedisConfig) workflow.RedisConfig {
	if override.Addr != "" {
		return override
	}
	return base
}

func mergeVars(base, override map[string]any) map[string]any {
	if base == nil && override == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

func mergeValidate(base, override []workflow.VarDef) []workflow.VarDef {
	if len(override) == 0 {
		return base
	}
	if len(base) == 0 {
		return override
	}

	seen := make(map[string]bool)
	var result []workflow.VarDef

	for _, v := range base {
		seen[v.Name] = true
		result = append(result, v)
	}
	for _, v := range override {
		if !seen[v.Name] {
			result = append(result, v)
		}
	}
	return result
}

func ApplyConfigToFile(cfg *ConfigFile, wf *workflow.Workflow, profileName string) error {
	merged := cfg.Config

	if profileName != "" {
		profile, err := cfg.GetProfile(profileName)
		if err != nil {
			return err
		}
		merged = MergeConfigs(merged, profile.Config)
	}

	if wf.Config.Vars["presets"] != nil {
		if presetNames, ok := wf.Config.Vars["presets"].([]any); ok {
			for _, pn := range presetNames {
				if name, ok := pn.(string); ok {
					preset, err := cfg.GetPreset(name)
					if err != nil {
						return err
					}
					merged = MergeConfigs(merged, preset.Config)
				}
			}
			delete(wf.Config.Vars, "presets")
		}
	}

	wf.Config = MergeConfigs(merged, wf.Config)

	return nil
}
