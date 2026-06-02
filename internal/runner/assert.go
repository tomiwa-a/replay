package runner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

type AssertionEngine struct {
	state map[string]any
}

func NewAssertionEngine(state map[string]any) *AssertionEngine {
	return &AssertionEngine{state: state}
}

func (e *AssertionEngine) Check(rule workflow.AssertRule, actual any) error {
	var data any
	if s, ok := actual.(string); ok && (strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")) {
		json.Unmarshal([]byte(s), &data)
	} else {
		data = actual
	}

	// Resolve variables in the expected Value if it's a string
	expectedValue := rule.Value
	if s, ok := rule.Value.(string); ok {
		rendered := template.Render(s, e.state)
		expectedValue = rendered
	}

	if rule.Path != "" && rule.Path != "$" {
		expr := rule.Path
		if !strings.HasPrefix(expr, "$") {
			expr = "$." + expr
		}
		// If path is "res.status" but data is the result map, we should map res. to $.
		if strings.HasPrefix(expr, "$.res.") {
			expr = strings.Replace(expr, "$.res.", "$.", 1)
		}

		exprObj, err := ParseJSONPath(expr, "", rule.Path)
		if err != nil {
			return err
		}
		values := exprObj.Get(data)
		if len(values) > 0 {
			actual = values[0]
		} else {
			actual = nil
		}
	} else {
		actual = data
	}

	// Type matching for numeric comparisons
	if rule.Op == "eq" || rule.Op == "==" || rule.Op == "=" {
		if aNum, ok1 := e.toFloat(actual); ok1 {
			if eNum, ok2 := e.toFloat(expectedValue); ok2 {
				if aNum == eNum {
					return nil
				}
			}
		}
	}

	actualVal := actual
	expectedVal := expectedValue

	switch rule.Op {
	case "eq", "==", "=":
		if reflect.DeepEqual(actualVal, expectedVal) {
			return nil
		}
		// Try string comparison fallback
		if fmt.Sprintf("%v", actualVal) == fmt.Sprintf("%v", expectedVal) {
			return nil
		}
		return fmt.Errorf("expected %v (%s) to equal %v", actualVal, rule.Path, expectedVal)
	case "ne", "!=", "<>":
		if reflect.DeepEqual(actualVal, expectedVal) {
			return fmt.Errorf("expected %v (%s) to not equal %v", actualVal, rule.Path, expectedVal)
		}
		// String comparison check
		if fmt.Sprintf("%v", actualVal) == fmt.Sprintf("%v", expectedVal) {
			return fmt.Errorf("expected %v (%s) to not equal %v", actualVal, rule.Path, expectedVal)
		}
	case "gt", ">":
		return e.compare(actualVal, expectedVal, "gt", rule.Path)
	case "lt", "<":
		return e.compare(actualVal, expectedVal, "lt", rule.Path)
	case "ge", ">=":
		return e.compare(actualVal, expectedVal, "ge", rule.Path)
	case "le", "<=":
		return e.compare(actualVal, expectedVal, "le", rule.Path)
	case "contains":
		sActual, ok1 := actualVal.(string)
		sExpected, ok2 := expectedVal.(string)
		if !ok1 || !ok2 {
			return fmt.Errorf("contains operator requires string values, got %T and %T", actualVal, expectedVal)
		}
		if !strings.Contains(sActual, sExpected) {
			return fmt.Errorf("expected %q (%s) to contain %q", sActual, rule.Path, sExpected)
		}
	case "in":
		// value 'in' array (where value is rendered expectedValue and array is actual)
		v := reflect.ValueOf(actualVal)
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return fmt.Errorf("'in' operator requires actual value to be a slice or array, got %T at %s", actualVal, rule.Path)
		}
		found := false
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(v.Index(i).Interface(), expectedVal) || fmt.Sprintf("%v", v.Index(i).Interface()) == fmt.Sprintf("%v", expectedVal) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected %v to be in %v at %s", expectedVal, actualVal, rule.Path)
		}
	case "not_null":
		if actualVal == nil {
			return fmt.Errorf("expected %s to be not null", rule.Path)
		}
	default:
		return fmt.Errorf("unsupported operator: %s", rule.Op)
	}
	return nil
}

func (e *AssertionEngine) compare(actual, expected any, op, path string) error {
	v1, ok1 := e.toFloat(actual)
	v2, ok2 := e.toFloat(expected)
	if !ok1 || !ok2 {
		return fmt.Errorf("%s operator requires numeric values, got %T and %T at %s", op, actual, expected, path)
	}

	switch op {
	case "gt":
		if !(v1 > v2) {
			return fmt.Errorf("expected %v (%s) to be greater than %v", v1, path, v2)
		}
	case "lt":
		if !(v1 < v2) {
			return fmt.Errorf("expected %v (%s) to be less than %v", v1, path, v2)
		}
	case "ge":
		if !(v1 >= v2) {
			return fmt.Errorf("expected %v (%s) to be greater than or equal to %v", v1, path, v2)
		}
	case "le":
		if !(v1 <= v2) {
			return fmt.Errorf("expected %v (%s) to be less than or equal to %v", v1, path, v2)
		}
	}
	return nil
}

func (e *AssertionEngine) toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case string:
		var f float64
		if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}
