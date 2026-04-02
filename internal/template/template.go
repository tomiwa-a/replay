package template

import (
	"fmt"
	"regexp"
)

var re = regexp.MustCompile(`{{\s*([\w\.-]+)\s*}}`)

func Render(input string, data map[string]any) string {
	return re.ReplaceAllStringFunc(input, func(m string) string {
		match := re.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		key := match[1]
		if val, ok := data[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return m
	})
}
