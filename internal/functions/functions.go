package functions

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/aymerick/raymond"
)

func toFloat(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}

func formatNum(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

func Register() {
	// String functions
	raymond.RegisterHelper("upper", func(s string) string {
		return strings.ToUpper(s)
	})
	raymond.RegisterHelper("lower", func(s string) string {
		return strings.ToLower(s)
	})
	raymond.RegisterHelper("upperFirst", func(s string) string {
		if s == "" {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	})
	raymond.RegisterHelper("lowerFirst", func(s string) string {
		if s == "" {
			return s
		}
		return strings.ToLower(s[:1]) + s[1:]
	})
	raymond.RegisterHelper("trim", func(s string) string {
		return strings.TrimSpace(s)
	})
	raymond.RegisterHelper("replace", func(s, old, new string) string {
		return strings.ReplaceAll(s, old, new)
	})
	raymond.RegisterHelper("contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})
	raymond.RegisterHelper("startsWith", func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	})
	raymond.RegisterHelper("endsWith", func(s, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	})
	raymond.RegisterHelper("padLeft", func(s string, width int, pad string) string {
		if pad == "" {
			pad = " "
		}
		for len(s) < width {
			s = pad[:1] + s
		}
		return s
	})
	raymond.RegisterHelper("padRight", func(s string, width int, pad string) string {
		if pad == "" {
			pad = " "
		}
		for len(s) < width {
			s = s + pad[:1]
		}
		return s
	})
	raymond.RegisterHelper("truncate", func(s string, maxLen int) string {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen-3] + "..."
	})
	raymond.RegisterHelper("repeat", func(s string, count int) string {
		return strings.Repeat(s, count)
	})
	raymond.RegisterHelper("split", func(s, sep string) []string {
		return strings.Split(s, sep)
	})
	raymond.RegisterHelper("join", func(arr []any, sep string) string {
		parts := make([]string, len(arr))
		for i, v := range arr {
			parts[i] = fmt.Sprintf("%v", v)
		}
		return strings.Join(parts, sep)
	})
	raymond.RegisterHelper("regexMatch", func(pattern, s string) bool {
		matched, err := regexp.MatchString(pattern, s)
		if err != nil {
			return false
		}
		return matched
	})
	raymond.RegisterHelper("regexFind", func(pattern, s string) string {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return ""
		}
		return re.FindString(s)
	})

	// Math functions (accept any, convert to float64)
	raymond.RegisterHelper("add", func(a, b any) string {
		return formatNum(toFloat(a) + toFloat(b))
	})
	raymond.RegisterHelper("sub", func(a, b any) string {
		return formatNum(toFloat(a) - toFloat(b))
	})
	raymond.RegisterHelper("mul", func(a, b any) string {
		return formatNum(toFloat(a) * toFloat(b))
	})
	raymond.RegisterHelper("div", func(a, b any) string {
		d := toFloat(b)
		if d == 0 {
			return "0"
		}
		return formatNum(toFloat(a) / d)
	})
	raymond.RegisterHelper("mod", func(a, b any) string {
		d := toFloat(b)
		if d == 0 {
			return "0"
		}
		return formatNum(math.Mod(toFloat(a), d))
	})
	raymond.RegisterHelper("min", func(a, b any) string {
		af, bf := toFloat(a), toFloat(b)
		if af < bf {
			return formatNum(af)
		}
		return formatNum(bf)
	})
	raymond.RegisterHelper("max", func(a, b any) string {
		af, bf := toFloat(a), toFloat(b)
		if af > bf {
			return formatNum(af)
		}
		return formatNum(bf)
	})
	raymond.RegisterHelper("round", func(a any) string {
		return formatNum(math.Round(toFloat(a)))
	})
	raymond.RegisterHelper("abs", func(a any) string {
		return formatNum(math.Abs(toFloat(a)))
	})
	raymond.RegisterHelper("ceil", func(a any) string {
		return formatNum(math.Ceil(toFloat(a)))
	})
	raymond.RegisterHelper("floor", func(a any) string {
		return formatNum(math.Floor(toFloat(a)))
	})

	// Date functions
	raymond.RegisterHelper("now", func() string {
		return time.Now().UTC().Format(time.RFC3339)
	})
	raymond.RegisterHelper("nowUnix", func() int64 {
		return time.Now().Unix()
	})
	raymond.RegisterHelper("addMinutes", func(t string, minutes int) string {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return t
		}
		return parsed.Add(time.Duration(minutes) * time.Minute).Format(time.RFC3339)
	})
	raymond.RegisterHelper("addHours", func(t string, hours int) string {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return t
		}
		return parsed.Add(time.Duration(hours) * time.Hour).Format(time.RFC3339)
	})
	raymond.RegisterHelper("addDays", func(t string, days int) string {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return t
		}
		return parsed.Add(time.Duration(days) * 24 * time.Hour).Format(time.RFC3339)
	})
	raymond.RegisterHelper("formatDate", func(t string, layout string) string {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return t
		}
		return parsed.Format(layout)
	})
	raymond.RegisterHelper("parseDate", func(s string, layout string) string {
		parsed, err := time.Parse(layout, s)
		if err != nil {
			return s
		}
		return parsed.Format(time.RFC3339)
	})
	raymond.RegisterHelper("dateSub", func(a, b string) float64 {
		tA, errA := time.Parse(time.RFC3339, a)
		tB, errB := time.Parse(time.RFC3339, b)
		if errA != nil || errB != nil {
			return 0
		}
		return tA.Sub(tB).Seconds()
	})

	// JSON/Object functions (return SafeString to avoid HTML escaping)
	raymond.RegisterHelper("jsonStringify", func(v any) raymond.SafeString {
		b, err := json.Marshal(v)
		if err != nil {
			return raymond.SafeString("")
		}
		return raymond.SafeString(string(b))
	})
	raymond.RegisterHelper("jsonParse", func(s string) any {
		var v any
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return nil
		}
		return v
	})
	raymond.RegisterHelper("jsonPick", func(obj any, path string) any {
		m, ok := obj.(map[string]any)
		if !ok {
			return nil
		}
		parts := strings.Split(strings.TrimPrefix(path, "$."), ".")
		var current any = m
		for _, part := range parts {
			cm, ok := current.(map[string]any)
			if !ok {
				return nil
			}
			current, ok = cm[part]
			if !ok {
				return nil
			}
		}
		return current
	})
	raymond.RegisterHelper("jsonKeys", func(obj any) []string {
		m, ok := obj.(map[string]any)
		if !ok {
			return nil
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	})
	raymond.RegisterHelper("jsonValues", func(obj any) []any {
		m, ok := obj.(map[string]any)
		if !ok {
			return nil
		}
		vals := make([]any, 0, len(m))
		for _, v := range m {
			vals = append(vals, v)
		}
		return vals
	})
	raymond.RegisterHelper("object", func(pairs ...any) map[string]any {
		result := make(map[string]any)
		for i := 0; i < len(pairs)-1; i += 2 {
			if key, ok := pairs[i].(string); ok {
				result[key] = pairs[i+1]
			}
		}
		return result
	})
	raymond.RegisterHelper("merge", func(a, b any) map[string]any {
		result := make(map[string]any)
		if mA, ok := a.(map[string]any); ok {
			for k, v := range mA {
				result[k] = v
			}
		}
		if mB, ok := b.(map[string]any); ok {
			for k, v := range mB {
				result[k] = v
			}
		}
		return result
	})

	// Type conversion functions
	raymond.RegisterHelper("toInt", func(v any) int {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case string:
			var i int
			fmt.Sscanf(n, "%d", &i)
			return i
		default:
			return 0
		}
	})
	raymond.RegisterHelper("toFloat", func(v any) float64 {
		switch n := v.(type) {
		case int:
			return float64(n)
		case int64:
			return float64(n)
		case float64:
			return n
		case string:
			var f float64
			fmt.Sscanf(n, "%f", &f)
			return f
		default:
			return 0
		}
	})
	raymond.RegisterHelper("toString", func(v any) string {
		return fmt.Sprintf("%v", v)
	})
	raymond.RegisterHelper("len", func(v any) int {
		switch val := v.(type) {
		case string:
			return len(val)
		case []any:
			return len(val)
		case map[string]any:
			return len(val)
		default:
			return 0
		}
	})
}
