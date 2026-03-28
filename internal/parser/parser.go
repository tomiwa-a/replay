package parser

import (
	"bytes"
	"fmt"
	"os"

	"github.com/replay/replay/internal/workflow"
	"gopkg.in/yaml.v3"
)

const defaultVersion = "v0.1"

func LoadFromFile(path string) (workflow.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return workflow.Workflow{}, fmt.Errorf("read workflow file %q: %w", path, err)
	}

	wf, err := LoadFromBytes(data)
	if err != nil {
		return workflow.Workflow{}, fmt.Errorf("parse workflow file %q: %w", path, err)
	}

	return wf, nil
}

func LoadFromBytes(data []byte) (workflow.Workflow, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var wf workflow.Workflow
	if err := decoder.Decode(&wf); err != nil {
		return workflow.Workflow{}, err
	}

	if wf.Version == "" {
		wf.Version = defaultVersion
	}

	return wf, nil
}
