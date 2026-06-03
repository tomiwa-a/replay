package functions

import (
	"testing"

	"github.com/aymerick/raymond"
)

func init() {
	Register()
}

func TestStringUpper(t *testing.T) {
	res, err := raymond.Render("{{ upper 'hello' }}", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res != "HELLO" {
		t.Errorf("expected HELLO, got %s", res)
	}
}

func TestStringLower(t *testing.T) {
	res, _ := raymond.Render("{{ lower 'HELLO' }}", nil)
	if res != "hello" {
		t.Errorf("expected hello, got %s", res)
	}
}

func TestStringUpperFirst(t *testing.T) {
	res, _ := raymond.Render("{{ upperFirst 'hello' }}", nil)
	if res != "Hello" {
		t.Errorf("expected Hello, got %s", res)
	}
}

func TestStringTrim(t *testing.T) {
	res, _ := raymond.Render("{{ trim '  hello  ' }}", nil)
	if res != "hello" {
		t.Errorf("expected hello, got %s", res)
	}
}

func TestStringReplace(t *testing.T) {
	res, _ := raymond.Render("{{ replace 'hello world' 'world' 'there' }}", nil)
	if res != "hello there" {
		t.Errorf("expected 'hello there', got %s", res)
	}
}

func TestStringContains(t *testing.T) {
	res, _ := raymond.Render("{{ contains 'hello world' 'world' }}", nil)
	if res != "true" {
		t.Errorf("expected true, got %s", res)
	}
}

func TestStringStartsWith(t *testing.T) {
	res, _ := raymond.Render("{{ startsWith 'hello' 'hel' }}", nil)
	if res != "true" {
		t.Errorf("expected true, got %s", res)
	}
}

func TestStringEndsWith(t *testing.T) {
	res, _ := raymond.Render("{{ endsWith 'hello' 'llo' }}", nil)
	if res != "true" {
		t.Errorf("expected true, got %s", res)
	}
}

func TestStringPadLeft(t *testing.T) {
	res, _ := raymond.Render("{{ padLeft '42' 5 '0' }}", nil)
	if res != "00042" {
		t.Errorf("expected 00042, got %s", res)
	}
}

func TestStringPadRight(t *testing.T) {
	res, _ := raymond.Render("{{ padRight 'hi' 5 '.' }}", nil)
	if res != "hi..." {
		t.Errorf("expected hi..., got %s", res)
	}
}

func TestStringTruncate(t *testing.T) {
	res, _ := raymond.Render("{{ truncate 'hello world' 8 }}", nil)
	if res != "hello..." {
		t.Errorf("expected 'hello...', got %s", res)
	}
}

func TestStringRepeat(t *testing.T) {
	res, _ := raymond.Render("{{ repeat 'ab' 3 }}", nil)
	if res != "ababab" {
		t.Errorf("expected ababab, got %s", res)
	}
}

func TestStringSplit(t *testing.T) {
	data := map[string]any{"items": "a,b,c"}
	res, _ := raymond.Render("{{#each (split items ',')}}{{this}} {{/each}}", data)
	if res != "a b c " {
		t.Errorf("expected 'a b c ', got %s", res)
	}
}

func TestStringRegexMatch(t *testing.T) {
	res, _ := raymond.Render("{{ regexMatch '^[a-z]+$' 'hello' }}", nil)
	if res != "true" {
		t.Errorf("expected true, got %s", res)
	}
}

func TestStringRegexFind(t *testing.T) {
	res, _ := raymond.Render("{{ regexFind '\\d+' 'abc123def' }}", nil)
	if res != "123" {
		t.Errorf("expected 123, got %s", res)
	}
}

func TestMathAdd(t *testing.T) {
	res, _ := raymond.Render("{{ add 3 4 }}", nil)
	if res != "7" {
		t.Errorf("expected 7, got %s", res)
	}
}

func TestMathSub(t *testing.T) {
	res, _ := raymond.Render("{{ sub 10 3 }}", nil)
	if res != "7" {
		t.Errorf("expected 7, got %s", res)
	}
}

func TestMathMul(t *testing.T) {
	res, _ := raymond.Render("{{ mul 3 4 }}", nil)
	if res != "12" {
		t.Errorf("expected 12, got %s", res)
	}
}

func TestMathDiv(t *testing.T) {
	res, _ := raymond.Render("{{ div 10 2 }}", nil)
	if res != "5" {
		t.Errorf("expected 5, got %s", res)
	}
}

func TestMathDivByZero(t *testing.T) {
	res, _ := raymond.Render("{{ div 10 0 }}", nil)
	if res != "0" {
		t.Errorf("expected 0, got %s", res)
	}
}

func TestMathMod(t *testing.T) {
	res, _ := raymond.Render("{{ mod 10 3 }}", nil)
	if res != "1" {
		t.Errorf("expected 1, got %s", res)
	}
}

func TestMathMin(t *testing.T) {
	res, _ := raymond.Render("{{ min 3 7 }}", nil)
	if res != "3" {
		t.Errorf("expected 3, got %s", res)
	}
}

func TestMathMax(t *testing.T) {
	res, _ := raymond.Render("{{ max 3 7 }}", nil)
	if res != "7" {
		t.Errorf("expected 7, got %s", res)
	}
}

func TestMathRound(t *testing.T) {
	res, _ := raymond.Render("{{ round 3.7 }}", nil)
	if res != "4" {
		t.Errorf("expected 4, got %s", res)
	}
}

func TestMathAbs(t *testing.T) {
	res, _ := raymond.Render("{{ abs -5 }}", nil)
	if res != "5" {
		t.Errorf("expected 5, got %s", res)
	}
}

func TestMathCeil(t *testing.T) {
	res, _ := raymond.Render("{{ ceil 3.2 }}", nil)
	if res != "4" {
		t.Errorf("expected 4, got %s", res)
	}
}

func TestMathFloor(t *testing.T) {
	res, _ := raymond.Render("{{ floor 3.9 }}", nil)
	if res != "3" {
		t.Errorf("expected 3, got %s", res)
	}
}

func TestDateNow(t *testing.T) {
	res, err := raymond.Render("{{ now }}", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestDateNowUnix(t *testing.T) {
	res, err := raymond.Render("{{ nowUnix }}", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res == "" {
		t.Error("expected non-empty unix timestamp")
	}
}

func TestDateAddMinutes(t *testing.T) {
	data := map[string]any{"ts": "2026-01-01T00:00:00Z"}
	res, _ := raymond.Render("{{ addMinutes ts 30 }}", data)
	if res != "2026-01-01T00:30:00Z" {
		t.Errorf("expected 2026-01-01T00:30:00Z, got %s", res)
	}
}

func TestDateAddHours(t *testing.T) {
	data := map[string]any{"ts": "2026-01-01T00:00:00Z"}
	res, _ := raymond.Render("{{ addHours ts 5 }}", data)
	if res != "2026-01-01T05:00:00Z" {
		t.Errorf("expected 2026-01-01T05:00:00Z, got %s", res)
	}
}

func TestDateAddDays(t *testing.T) {
	data := map[string]any{"ts": "2026-01-01T00:00:00Z"}
	res, _ := raymond.Render("{{ addDays ts 7 }}", data)
	if res != "2026-01-08T00:00:00Z" {
		t.Errorf("expected 2026-01-08T00:00:00Z, got %s", res)
	}
}

func TestDateFormatDate(t *testing.T) {
	data := map[string]any{"ts": "2026-01-15T14:30:00Z"}
	res, _ := raymond.Render("{{ formatDate ts '2006-01-02' }}", data)
	if res != "2026-01-15" {
		t.Errorf("expected 2026-01-15, got %s", res)
	}
}

func TestDateParseDate(t *testing.T) {
	res, _ := raymond.Render("{{ parseDate '15/01/2026' '02/01/2006' }}", nil)
	if res != "2026-01-15T00:00:00Z" {
		t.Errorf("expected 2026-01-15T00:00:00Z, got %s", res)
	}
}

func TestDateSub(t *testing.T) {
	data := map[string]any{
		"a": "2026-01-02T00:00:00Z",
		"b": "2026-01-01T00:00:00Z",
	}
	res, _ := raymond.Render("{{ dateSub a b }}", data)
	if res != "86400" {
		t.Errorf("expected 86400, got %s", res)
	}
}

func TestJSONStringify(t *testing.T) {
	data := map[string]any{"obj": map[string]any{"a": 1}}
	res, _ := raymond.Render("{{ jsonStringify obj }}", data)
	if res != `{"a":1}` {
		t.Errorf("expected {\"a\":1}, got %s", res)
	}
}

func TestJSONParse(t *testing.T) {
	res, _ := raymond.Render("{{ jsonParse '{\"x\":1}' }}", nil)
	if res == "" {
		t.Error("expected non-empty result")
	}
}

func TestJSONPick(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"age":  30,
		},
	}
	res, _ := raymond.Render("{{ jsonPick user 'name' }}", data)
	if res != "alice" {
		t.Errorf("expected alice, got %s", res)
	}
}

func TestJSONKeys(t *testing.T) {
	data := map[string]any{
		"obj": map[string]any{"a": 1, "b": 2},
	}
	res, err := raymond.Render("{{#each (jsonKeys obj)}}{{this}} {{/each}}", data)
	if err != nil {
		t.Fatal(err)
	}
	if res == "" {
		t.Error("expected non-empty keys")
	}
}

func TestJSONMerge(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{"x": 1},
		"b": map[string]any{"y": 2},
	}
	res, err := raymond.Render("{{ jsonStringify (merge a b) }}", data)
	if err != nil {
		t.Fatal(err)
	}
	if res != `{"x":1,"y":2}` {
		t.Errorf("expected {\"x\":1,\"y\":2}, got %s", res)
	}
}

func TestTypeToInt(t *testing.T) {
	res, _ := raymond.Render("{{ toInt 3.7 }}", nil)
	if res != "3" {
		t.Errorf("expected 3, got %s", res)
	}
}

func TestTypeToFloat(t *testing.T) {
	res, _ := raymond.Render("{{ toFloat 42 }}", nil)
	if res != "42" {
		t.Errorf("expected 42, got %s", res)
	}
}

func TestTypeToString(t *testing.T) {
	res, _ := raymond.Render("{{ toString 42 }}", nil)
	if res != "42" {
		t.Errorf("expected 42, got %s", res)
	}
}

func TestTypeLen(t *testing.T) {
	res, _ := raymond.Render("{{ len 'hello' }}", nil)
	if res != "5" {
		t.Errorf("expected 5, got %s", res)
	}
}

func TestTypeLenArray(t *testing.T) {
	data := map[string]any{"arr": []any{1, 2, 3}}
	res, _ := raymond.Render("{{ len arr }}", data)
	if res != "3" {
		t.Errorf("expected 3, got %s", res)
	}
}

func TestCombinedFunctions(t *testing.T) {
	data := map[string]any{"name": "alice"}
	res, _ := raymond.Render("{{ upper (trim name) }}", data)
	if res != "ALICE" {
		t.Errorf("expected ALICE, got %s", res)
	}
}

func TestFunctionsInWorkflowContext(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"id":   42,
		},
		"items": []any{"a", "b", "c"},
	}
	res, _ := raymond.Render("Hello {{ upperFirst user.name }}! You have {{ len items }} items.", data)
	if res != "Hello Alice! You have 3 items." {
		t.Errorf("expected 'Hello Alice! You have 3 items.', got %s", res)
	}
}
