package runner

import (
	"fmt"

	"github.com/ohler55/ojg/jp"
)

// ParseJSONPath parses a JSONPath expression and returns a descriptive error if it fails.
// The stepName and varName are used to provide context in the error message.
func ParseJSONPath(path, stepName, varName string) (jp.Expr, error) {
	expr, err := jp.ParseString(path)
	if err != nil {
		if stepName != "" && varName != "" {
			return nil, fmt.Errorf("step %q: invalid JSONPath %q for variable %q: %w", stepName, path, varName, err)
		}
		if stepName != "" {
			return nil, fmt.Errorf("step %q: invalid JSONPath %q: %w", stepName, path, err)
		}
		return nil, fmt.Errorf("invalid JSONPath %q: %w", path, err)
	}
	return expr, nil
}