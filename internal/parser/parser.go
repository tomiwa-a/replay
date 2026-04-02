package parser

import (
	"bytes"
	"fmt"
	"os"

	"github.com/replay/replay/internal/workflow"
	"gopkg.in/yaml.v3"
)

const defaultVersion = "v0.1"

func LoadFromFile(path string) ([]workflow.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file %q: %w", path, err)
	}

	return LoadFromBytes(data)
}

func LoadFromBytes(data []byte) ([]workflow.Workflow, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var wfs []workflow.Workflow
	for {
		var wf workflow.Workflow
		if err := decoder.Decode(&wf); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		if wf.Version == "" {
			wf.Version = defaultVersion
		}
		wfs = append(wfs, wf)
	}

	return wfs, nil
}
