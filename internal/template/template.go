package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aymerick/raymond"
)

var varPattern = regexp.MustCompile(`\{\{\s*([^}]+?)\s*\}\}`)

func Render(input string, data map[string]any) string {
	res, err := raymond.Render(input, data)
	if err == nil {
		return res
	}
	return renderWithDefaults(input, data)
}

func renderWithDefaults(input string, data map[string]any) string {
	return varPattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := varPattern.FindStringSubmatch(match)
		if len(groups) < 2 {
			return match
		}
		varName := strings.TrimSpace(groups[1])
		if val, ok := getNestedValue(data, varName); ok {
			return toString(val)
		}
		return ""
	})
}

func getNestedValue(data map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = data
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		current = val
	}
	return current, true
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
