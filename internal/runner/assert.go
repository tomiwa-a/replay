package runner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/ohler55/ojg/jp"
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

	if rule.Path != "" && rule.Path != "$" {
		expr := rule.Path
		if !strings.HasPrefix(expr, "$") {
			expr = "$." + expr
		}
		if p, err := jp.ParseString(expr); err == nil {
			values := p.Get(data)
			if len(values) > 0 {
				actual = values[0]
			} else {
				actual = nil
			}
		}
	} else {
		actual = data
	}

	switch rule.Op {
	case "eq":
		if !reflect.DeepEqual(actual, rule.Value) {
			return fmt.Errorf("expected %v to equal %v", actual, rule.Value)
		}
	case "ne":
		if reflect.DeepEqual(actual, rule.Value) {
			return fmt.Errorf("expected %v to not equal %v", actual, rule.Value)
		}
	case "gt":
		return e.compare(actual, rule.Value, "gt")
	case "lt":
		return e.compare(actual, rule.Value, "lt")
	case "contains":
		sActual, ok1 := actual.(string)
		sExpected, ok2 := rule.Value.(string)
		if !ok1 || !ok2 {
			return fmt.Errorf("contains operator requires string values, got %T and %T", actual, rule.Value)
		}
		if !strings.Contains(sActual, sExpected) {
			return fmt.Errorf("expected %q to contain %q", sActual, sExpected)
		}
	case "not_null":
		if actual == nil {
			return fmt.Errorf("expected value to be not null")
		}
	default:
		return fmt.Errorf("unsupported operator: %s", rule.Op)
	}
	return nil
}

func (e *AssertionEngine) compare(actual, expected any, op string) error {
	v1, ok1 := e.toFloat(actual)
	v2, ok2 := e.toFloat(expected)
	if !ok1 || !ok2 {
		return fmt.Errorf("%s operator requires numeric values, got %T and %T", op, actual, expected)
	}

	switch op {
	case "gt":
		if !(v1 > v2) {
			return fmt.Errorf("expected %v to be greater than %v", v1, v2)
		}
	case "lt":
		if !(v1 < v2) {
			return fmt.Errorf("expected %v to be less than %v", v1, v2)
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
	}
	return 0, false
}
