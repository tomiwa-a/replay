# Template Reference

Variable interpolation, JSONPath extraction, and 45 built-in functions for replay workflows. Always confirm against https://docs.ellomas.com/replay/guides/templating-and-functions.

---

## 1. Variable Interpolation `{{ var }}`

Any string field in a step can reference state variables:

```yaml
url: /users/{{ user_id }}
headers:
  Authorization: Bearer {{ token }}
body:
  email: "{{ user_email }}"
message: "Status: {{ status_code }}"
```

Variables are resolved at execution time from the shared **state bag**.

---

## 2. Variable Sources

| Source | Example | Description |
|--------|---------|-------------|
| Step `extract` | `{{ token }}` | Variables extracted via `extract:` blocks |
| Workflow `config.vars` | `{{ region }}` | Convenience variables defined in workflow config |
| Environment variables | `{{ DATABASE_URL }}` | Shell env vars loaded at startup into global scope |
| Loop item | `{{ item }}` | Current element in a `loop` step |
| Loop index | `{{ index }}` | Zero-based loop index |
| Include `with` | `{{ env }}` | Parameters passed via `include` → `with` block |
| Call `with` | `{{ email }}` | Parameters passed via `call` → `with` block |
| Built-in | `{{ nowUnix }}` | `nowUnix`, `now`, `random` (see functions below) |

### Extraction via `extract:` block

```yaml
extract:
  token: $.data.token           # JSONPath into response
  user_name: data.user.name     # Shorthand (no $ prefix)
  first_item: data.items[0]     # Array index access
```

### Nested Access

Dot notation for nested values:

```yaml
message: "Hello {{ user.profile.first_name }}"
url: /organizations/{{ org.id }}/members/{{ member.id }}
```

---

## 3. JSONPath Reference

Used in `extract:` blocks and `assert:` path fields.

| Expression | Description | Example |
|------------|-------------|---------|
| `$` | Root object | `$` |
| `$.data` | Nested field | `$.data.token` |
| `data.name` | Shorthand (no `$` prefix) | `data.name` |
| `$.items[0]` | Array index | `$.data.users[0]` |
| `$.items[*].name` | Wildcard (all elements) | `$.data.items[*].id` |
| `res.header.Server[0]` | Response header | `res.header.Content-Type[0]` |
| `$.status` | HTTP status code | `$.status` |
| `res.status` | Status (explicit prefix) | `res.status` |

### Response Structure for Different Step Types

**HTTP:**
```json
{ "status": 200, "data": { ... }, "header": { ... } }
```

**Shell:**
```json
{ "stdout": "...", "stderr": "..." }
```

**DB (Postgres):**
```json
[{ "id": 1, "email": "..." }, { "id": 2, "email": "..." }]
```

**DB (Redis):**
```json
"value_string"
```

---

## 4. Assert Operators

Used in `assert:` blocks and `if` conditions.

| Operator | Aliases | Type | Example |
|----------|---------|------|---------|
| `eq` | `==`, `=` | Any | `["$.status", "eq", 200]` |
| `ne` | `!=`, `<>` | Any | `["$.data.role", "!=", "banned"]` |
| `gt` | `>` | Numeric | `["$.data.count", ">", 0]` |
| `lt` | `<` | Numeric | `["$.data.age", "<", 18]` |
| `ge` | `>=` | Numeric | `["$.data.score", ">=", 100]` |
| `le` | `<=` | Numeric | `["$.data.price", "<=", 50]` |
| `contains` | | String | `["$.data.email", "contains", "@"]` |
| `in` | | Array | `["$.data.role", "in", "admin,editor"]` |
| `not_null` | | Any | `["$.data.token", "not_null"]` |

---

## 5. Built-in Functions (45 total)

### 5.1 String Functions

| Function | Description | Example | Result |
|----------|-------------|---------|--------|
| `upper` | Convert to uppercase | `{{upper "hello"}}` | `HELLO` |
| `lower` | Convert to lowercase | `{{lower "HELLO"}}` | `hello` |
| `upperFirst` | Capitalize first letter | `{{upperFirst "hello"}}` | `Hello` |
| `lowerFirst` | Lowercase first letter | `{{lowerFirst "Hello"}}` | `hello` |
| `trim` | Trim whitespace | `{{trim " hi "}}` | `hi` |
| `replace` | Replace all occurrences | `{{replace "a-b-c" "-" "."}}` | `a.b.c` |
| `contains` | Check substring | `{{contains "hello" "ell"}}` | `true` |
| `startsWith` | Check prefix | `{{startsWith "hello" "he"}}` | `true` |
| `endsWith` | Check suffix | `{{endsWith "hello" "lo"}}` | `true` |
| `split` | Split string | `{{split "a,b,c" ","}}` | `["a","b","c"]` |
| `join` | Join array | `{{join (split "a,b" ",") "-"}}` | `a-b` |
| `truncate` | Truncate with ellipsis | `{{truncate "hello world" 8}}` | `hello...` |
| `padLeft` | Pad left | `{{padLeft "42" 5 "0"}}` | `00042` |
| `padRight` | Pad right | `{{padRight "42" 5 "."}}` | `42...` |
| `repeat` | Repeat string | `{{repeat "ab" 3}}` | `ababab` |
| `regexMatch` | Regex test | `{{regexMatch "^\\d+$" "123"}}` | `true` |
| `regexFind` | Regex find first | `{{regexFind "\\d+" "abc123def"}}` | `123` |

### 5.2 Math Functions

| Function | Description | Example | Result |
|----------|-------------|---------|--------|
| `add` | Addition | `{{add 5 3}}` | `8` |
| `sub` | Subtraction | `{{sub 10 4}}` | `6` |
| `mul` | Multiplication | `{{mul 3 4}}` | `12` |
| `div` | Division | `{{div 10 3}}` | `3.333` |
| `mod` | Modulo | `{{mod 10 3}}` | `1` |
| `min` | Minimum | `{{min 5 3}}` | `3` |
| `max` | Maximum | `{{max 5 3}}` | `5` |
| `round` | Round | `{{round 3.7}}` | `4` |
| `abs` | Absolute value | `{{abs -5}}` | `5` |
| `ceil` | Ceiling | `{{ceil 3.2}}` | `4` |
| `floor` | Floor | `{{floor 3.8}}` | `3` |

### 5.3 Date Functions

| Function | Description | Example | Result |
|----------|-------------|---------|--------|
| `now` | Current UTC (RFC3339) | `{{now}}` | `2026-06-03T12:00:00Z` |
| `nowUnix` | Current Unix timestamp | `{{nowUnix}}` | `1759507200` |
| `addMinutes` | Add minutes | `{{addMinutes "2026-01-01T00:00:00Z" 30}}` | `2026-01-01T00:30:00Z` |
| `addHours` | Add hours | `{{addHours "2026-01-01T00:00:00Z" 2}}` | `2026-01-01T02:00:00Z` |
| `addDays` | Add days | `{{addDays "2026-01-01T00:00:00Z" 7}}` | `2026-01-08T00:00:00Z` |
| `formatDate` | Format date | `{{formatDate "2026-01-01T00:00:00Z" "2006-01-02"}}` | `2026-01-01` |
| `parseDate` | Parse date string | `{{parseDate "2026-01-01" "2006-01-02"}}` | `2026-01-01T00:00:00Z` |
| `dateSub` | Date diff (seconds) | `{{dateSub "2026-01-02T00:00:00Z" "2026-01-01T00:00:00Z"}}` | `86400` |

### 5.4 JSON Functions

| Function | Description |
|----------|-------------|
| `jsonStringify` | Convert value to JSON string |
| `jsonParse` | Parse JSON string to value |
| `jsonPick` | Extract nested value by path |
| `jsonKeys` | Get object keys |
| `jsonValues` | Get object values |
| `object` | Build object from key-value pairs |
| `merge` | Merge two objects |

### 5.5 Type Conversion Functions

| Function | Description |
|----------|-------------|
| `toInt` | Convert to integer |
| `toFloat` | Convert to float |
| `toString` | Convert to string |
| `len` | Length of string, array, or object |

---

## 6. Function Composition

Functions can be nested:

```yaml
message: "{{ upper (trim user_name) }}"
message: "{{ add (toInt count) 1 }}"
message: "{{ join (split tags ",") ", " }}"
email: "qa-{{ nowUnix }}@example.com"
```

---

## 7. Variables in Assert Expected Values

Assert expected values can reference state variables:

```yaml
- name: set-expected
  type: shell
  command: 'echo "{\"expected_name\": \"Bob\"}"'
  extract:
    target_name: $.expected_name

- name: check-assertion
  type: shell
  command: 'echo "{\"name\": \"Bob\"}"'
  assert:
    - ["$.name", "eq", "{{ target_name }}"]
```

---

## 8. Variable Scope

Three-level scope hierarchy:

1. **Global** — Environment variables loaded at startup (`{{ DATABASE_URL }}`)
2. **Workflow** — Set via `config.vars` or workflow-level `extract`
3. **Step** — Set via `extract` within a step

Inner scopes inherit from outer scopes. Variables are cleaned up when their containing scope exits.
