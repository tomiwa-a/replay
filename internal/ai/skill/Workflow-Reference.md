# Workflow Reference

Full specification for the replay workflow DSL. Use this as the authoritative reference when generating or editing replay workflows. Always confirm against https://docs.ellomas.com/replay.

---

## 1. Root Structure

```yaml
name: my-workflow              # Required — workflow name
version: v0.1                  # Optional — DSL version (defaults to v0.1)
description: "..."             # Optional — human-readable description
config:                        # Optional — workflow-level configuration
  http:
    base_url: https://api.example.com
    headers:
      Accept: application/json
    debug: false
  postgres:
    dsn: postgres://user:pass@localhost:5432/db
  redis:
    addr: localhost:6379
  vars:
    region: us-east-1          # Convenience variables
include:                       # Optional — compile-time step injection
  - file: setup.yaml
    with:
      env: test
steps:                         # Required — ordered list of steps
  - name: step-name
    type: http
    ...
```

---

## 2. Step Types

### 2.1 `http` — HTTP Request

Makes an HTTP request and captures the response.

```yaml
- name: get-user
  type: http
  request:
    method: GET                 # GET, POST, PUT, PATCH, DELETE
    url: /users/1               # Relative (resolves against base_url) or absolute
    headers:                    # Optional. Per-step headers override config-level.
      Authorization: Bearer {{ token }}
    body:                       # Optional. Request body (map or raw string).
      email: user@example.com
      password: secret123
```

**Response structure:**
```json
{
  "status": 200,
  "data": { "id": 1, "email": "..." },
  "header": { "content-type": ["application/json"] }
}
```

**Config shorthand:**
```yaml
config:
  http:
    base_url: https://api.example.com  # Resolved for relative URLs
    headers:
      Accept: application/json         # Applied to every request
    debug: true                        # Enable request/response logging
```

**Assert + extract:**
```yaml
- name: login
  type: http
  request:
    method: POST
    url: /auth/login
    body:
      email: admin@example.com
      password: "{{ ADMIN_PASSWORD }}"
  extract:
    token: $.data.accessToken
    user_id: $.data.user.id
  assert:
    - ["$.status", "eq", 200]
    - ["$.data.accessToken", "not_null"]
```

### 2.2 `shell` — Shell Command

Executes shell commands and captures output.

```yaml
# Single command
- name: check-disk
  type: shell
  command: "df -h /"

# Multiple sequential commands
- name: setup
  type: shell
  commands:
    - "mkdir -p /tmp/test"
    - "echo config > /tmp/test/config.json"

# Parallel commands
- name: parallel-tasks
  type: shell
  parallel: true
  commands:
    - "sleep 3 && echo 'Task 1 done'"
    - "sleep 2 && echo 'Task 2 done'"

# With timeout and working directory
- name: build
  type: shell
  command: "make build"
  dir: /projects/my-app
  timeout: 30s
```

**Output variables:**
| Variable | Description |
|----------|-------------|
| `stdout` | Stdout of last command |
| `stderr` | Stderr of last command |
| `stdout_0`, `stdout_1`... | Per-command stdout (parallel mode) |
| `stderr_0`, `stderr_1`... | Per-command stderr (parallel mode) |

### 2.3 `db` — Database Operation

Executes PostgreSQL queries or Redis commands.

```yaml
# PostgreSQL
- name: query-users
  type: db
  query: "SELECT id, email FROM users WHERE status = 'active'"

- name: create-user
  type: db
  query: |
    INSERT INTO users (name, email)
    VALUES ('Alice', 'alice@example.com')
    RETURNING id;
  extract:
    new_id: $[0].id

# Redis
- name: set-cache
  type: db
  engine: redis
  command: ["SET", "session:1", "active"]

- name: get-cache
  type: db
  engine: redis
  command: ["GET", "session:1"]
  extract:
    session_status: $
  assert:
    - ["$", "eq", "active"]
```

**Postgres config:**
```yaml
config:
  postgres:
    dsn: postgres://user:password@localhost:5432/mydb
```

**Redis config:**
```yaml
config:
  redis:
    addr: localhost:6379
```

### 2.4 `print` — Print Message

```yaml
- name: log-status
  type: print
  message: "User {{ user_id }} has status {{ status }}"
```

### 2.5 `if` — Conditional Branching

```yaml
- name: check-status
  type: if
  condition: ["status_code", "==", 200]
  then:
    - name: log-success
      type: print
      message: "Request succeeded"
  else:
    - name: log-failure
      type: print
      message: "Request failed with status {{ status_code }}"
```

**Condition format:** `[value_or_var, operator, expected_value]`

| Operator | Aliases | Type | Description |
|----------|---------|------|-------------|
| `eq` | `==`, `=` | Any | Equal (type-aware for numbers) |
| `ne` | `!=`, `<>` | Any | Not equal |
| `gt` | `>` | Numeric | Greater than |
| `lt` | `<` | Numeric | Less than |
| `ge` | `>=` | Numeric | Greater than or equal |
| `le` | `<=` | Numeric | Less than or equal |
| `contains` | | String | Target contains substring |
| `in` | | Array | Value exists in array |

**Nested if:**
```yaml
- name: nested-logic
  type: if
  condition: ["authenticated", "==", true]
  then:
    - name: check-role
      type: if
      condition: ["role", "==", "admin"]
      then:
        - name: grant-admin
          type: print
          message: "Admin access granted"
```

### 2.6 `loop` — Iteration

```yaml
- name: process-items
  type: loop
  foreach: item_list, item     # list_var, element_var
  steps:
    - name: print-item
      type: print
      message: "[{{ index }}] {{ item.name }}"
```

**`foreach` format:** `list_variable, item_variable`

**Loop variables:**
| Variable | Description |
|----------|-------------|
| `item` | Current element (name matches second part of `foreach`) |
| `index` | Zero-based loop index |

**Extracting arrays for loops:**
```yaml
- name: fetch-staff
  type: http
  request:
    method: GET
    url: /staff
  extract:
    staff_list: $.data.items

- name: list-each
  type: loop
  foreach: staff_list, staff
  steps:
    - name: print-member
      type: print
      message: "{{ staff.firstName }} {{ staff.lastName }}"
```

### 2.7 `call` — Runtime Workflow Composition

```yaml
- name: authenticate
  type: call
  file: auth.yaml
  with:
    email: qa@example.com
    password: "{{ QA_PASSWORD }}"
  returns:
    - token
    - session_id
```

**Target specific step:**
```yaml
- name: reset-db
  type: call
  file: db-operations.yaml
  target: truncate-tables
```

---

## 3. `include` — Compile-Time Injection

```yaml
name: full-test
include:
  - file: setup.yaml
    with:
      env: test
steps:
  - name: run-test
    type: http
    request:
      method: GET
      url: /health
```

**Include vs Call:**

| | Include | Call |
|--|---------|------|
| **When** | Parse time (before execution) | Runtime (mid-execution) |
| **State** | Shares full state | Supports `returns` isolation |
| **Parameters** | `{{ var }}` substitution | `with` block + `{{ var }}` |
| **Targeting** | Cannot target — all steps injected | Can target a specific step |

---

## 4. Extract

Maps JSONPath expressions to state bag variable names.

```yaml
extract:
  token: $.data.token            # Nested field
  user_id: data.user.id          # Shorthand (no $ prefix)
  first_item: data.items[0]      # Array index
  all_emails: data.items[*].email # Array wildcard
  server_header: res.header.Server[0]  # Response header
```

---

## 5. Assert

Validates step outputs. Each rule: `[path, operator, expected_value]`

```yaml
assert:
  - ["$.status", "eq", 200]
  - ["$.data.email", "not_null"]
  - ["$.data.count", "gt", 0]
  - ["$.data.role", "in", "admin,editor"]
```

Also supported as object form:
```yaml
assert:
  - path: $.status
    op: eq
    value: 200
```

---

## 6. Error Handling

```yaml
- name: risky-step
  type: shell
  command: "might-fail"
  ignore_error: true            # Continue even if this step fails

- name: this-still-runs
  type: print
  message: "Continues regardless"
```

**Cleanup** — remove sensitive variables from state:
```yaml
- name: temp-data
  type: shell
  command: 'echo "{\"temp\": \"secret\"}"'
  extract:
    temp_key: $.temp
  cleanup:
    - temp_key
```

---

## 7. Configuration Profiles

```yaml
# replay.yaml
version: "1"
config:
  http:
    base_url: http://localhost:3000
  postgres:
    dsn: postgres://localhost:5432/devdb

profiles:
  staging:
    config:
      http:
        base_url: https://staging.example.com
      postgres:
        dsn: postgres://deploy:pass@staging-db:5432/stagingdb
  production:
    config:
      http:
        base_url: https://api.example.com
```

Usage:
```bash
replay run workflow.yaml --profile staging
```

Merge order (later overrides earlier):
1. Config file default (`config` section)
2. Selected profile (`profiles.<name>.config`)
3. Workflow file's own `config` section
