package template

import (
	"github.com/aymerick/raymond"
)

func Render(input string, data map[string]any) string {
	res, err := raymond.Render(input, data)
	if err != nil {
		// If it's an undefined variable error, we want to keep the template as-is for those variables
		// but replace the ones we can. Let's try a different approach.
		// For now, return the input on any error to maintain current behavior for most cases
		return input
	}
	return res
}
