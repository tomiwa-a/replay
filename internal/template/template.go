package template

import (
	"github.com/aymerick/raymond"
)

func Render(input string, data map[string]any) string {
	res, err := raymond.Render(input, data)
	if err != nil {
		return input
	}
	return res
}
