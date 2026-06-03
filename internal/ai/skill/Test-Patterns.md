# Test Patterns

Standard testing scenarios, folder organization conventions, workflow editing patterns, and CI/CD integration for replay. Always confirm against https://docs.ellomas.com/replay/guides/best-practices.

---

## 1. Folder Structure

Organise workflows by domain for maintainability and parallel execution:

```
tests/
├── auth/                       # Authentication & authorization
│   ├── login.yaml
│   ├── registration.yaml
│   └── password-reset.yaml
├── api/                        # API CRUD tests
│   ├── crud-users.yaml
│   ├── pagination.yaml
│   └── filtering.yaml
├── db/                         # Database operations
│   ├── seed.yaml
│   └── migrations.yaml
├── regression/                 # Full regression suite
│   ├── checkout.yaml
│   └── onboarding.yaml
├── smoke/                      # Quick health-check workflows
│   └── health.yaml
├── auth.yaml                   # Reusable auth workflow (called via `call`)
├── setup.yaml                  # Included setup steps
└── replay.yaml                 # Config profiles & presets
```

Run everything in parallel:
```bash
replay run tests/**/*.yaml --concurrency 4 --fail-fast
```

---

## 2. Happy Path (200/201)

```yaml
name: create-user-happy-path
config:
  http:
    base_url: "{{ BASE_URL }}"
steps:
  - name: create-user
    type: http
    request:
      method: POST
      url: /users
      headers:
        Content-Type: application/json
      body:
        email: "qa-{{ nowUnix }}@example.com"
        name: "Test User"
    extract:
      user_id: $.data.id
      user_email: $.data.email
    assert:
      - ["$.status", "eq", 201]
      - ["$.data.id", "not_null"]
      - ["$.data.email", "contains", "@"]
```

---

## 3. Validation Errors (422)

Test each validation rule independently.

```yaml
name: create-user-validation
steps:
  - name: missing-email
    type: http
    request:
      method: POST
      url: /users
      headers:
        Content-Type: application/json
      body:
        name: "No Email"
    assert:
      - ["$.status", "eq", 422]
      - ["$.data.error.code", "eq", "VALIDATION_ERROR"]
      - ["$.data.error.fields.email", "not_null"]
```

---

## 4. Authentication Failures (401/403)

```yaml
name: auth-failures
config:
  http:
    base_url: "{{ BASE_URL }}"
steps:
  # Missing token
  - name: no-token
    type: http
    request:
      method: GET
      url: /protected/resource
    assert:
      - ["$.status", "eq", 401]
      - ["$.data.error.code", "eq", "UNAUTHORIZED"]

  # Invalid token
  - name: bad-token
    type: http
    request:
      method: GET
      url: /protected/resource
      headers:
        Authorization: Bearer invalid-jwt
    assert:
      - ["$.status", "eq", 401"]
      - ["$.data.error.code", "eq", "INVALID_TOKEN"]
```

---

## 5. Multi-Step Chaining (Auth Flow)

```yaml
name: authenticated-crud
config:
  http:
    base_url: "{{ BASE_URL }}"
steps:
  - name: login
    type: http
    request:
      method: POST
      url: /auth/login
      body:
        email: "{{ TEST_USER_EMAIL }}"
        password: "{{ TEST_USER_PASSWORD }}"
    extract:
      token: $.data.token
    assert:
      - ["$.status", "eq", 200]
      - ["$.data.token", "not_null"]

  - name: create-item
    type: http
    request:
      method: POST
      url: /items
      headers:
        Authorization: Bearer {{ token }}
      body:
        title: "Test Item"
    extract:
      item_id: $.data.id
    assert:
      - ["$.status", "eq", 201]

  - name: get-item
    type: http
    request:
      method: GET
      url: /items/{{ item_id }}
      headers:
        Authorization: Bearer {{ token }}
    assert:
      - ["$.status", "eq", 200"]
      - ["$.data.title", "eq", "Test Item"]
```

---

## 6. Async Polling (Submit → Loop → Complete)

```yaml
name: async-score-request
steps:
  - name: submit-score
    type: http
    request:
      method: POST
      url: /scores/request
      body:
        score_id: "score-{{ nowUnix }}"
    extract:
      request_id: $.data.request_id
    assert:
      - ["$.status", "eq", 202]
      - ["$.data.request_id", "not_null"]

  - name: poll-result
    type: loop
    foreach: "1..30, attempt"
    steps:
      - name: check-score
        type: http
        request:
          method: GET
          url: /scores/request/{{ request_id }}
    condition: ["check_score_status", "==", "completed"]

  - name: verify-completion
    type: http
    request:
      method: GET
      url: /scores/request/{{ request_id }}
    assert:
      - ["$.status", "eq", 200]
      - ["$.data.status", "eq", "completed"]
```

---

## 7. Database Verification

```yaml
name: db-verify
steps:
  - name: create-user
    type: http
    request:
      method: POST
      url: /users
      body:
        email: "verify-{{ nowUnix }}@example.com"
    extract:
      created_email: $.data.email
    assert:
      - ["$.status", "eq", 201]

  - name: check-db
    type: db
    query: "SELECT id, email, status FROM users WHERE email = '{{ created_email }}'"
    assert:
      - ["$[0].id", "not_null"]
      - ["$[0].email", "eq", "{{ created_email }}"]
      - ["$[0].status", "eq", "active"]
```

---

## 8. Reusable Auth via `call`

**auth.yaml:**
```yaml
name: login
steps:
  - name: login
    type: http
    request:
      method: POST
      url: /auth/login
      body:
        email: "{{ email }}"
        password: "{{ password }}"
    extract:
      token: $.data.token
    returns: [token]
```

**test.yaml:**
```yaml
name: authenticated-test
steps:
  - name: authenticate
    type: call
    file: auth.yaml
    with:
      email: admin@example.com
      password: "{{ ADMIN_PASSWORD }}"
    returns:
      - token

  - name: use-token
    type: http
    request:
      method: GET
      url: /admin/dashboard
      headers:
        Authorization: Bearer {{ token }}
```

---

## 9. Setup → Test → Teardown Pattern

```yaml
name: managed-test
include:
  - file: seed-database.yaml
    with:
      env: test
steps:
  - name: run-tests
    type: http
    request:
      method: GET
      url: /users
    assert:
      - ["$.status", "eq", 200]

  - name: cleanup
    type: call
    file: cleanup.yaml
    with:
      env: test
```

---

## 10. CI/CD Integration

### GitHub Actions

```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: password
        options: --health-cmd pg_isready --health-interval 5s
    steps:
      - uses: actions/checkout@v4
      - name: Install Replay
        run: go install github.com/tomiwa-a/replay@latest
      - name: Run E2E workflows
        run: replay run tests/e2e/*.yaml --concurrency 4 --fail-fast
```

### Local Development

```bash
# Watch mode
replay run tests/auth/login.yaml --watch

# Parallel with debug
replay run tests/**/*.yaml --concurrency 4 --debug

# With profile
replay run smoke.yaml --profile staging

# Validate before run
replay validate workflow.yaml && replay run workflow.yaml
```

### Exit Codes
- `0` — All workflows passed
- `1` — One or more workflows failed

---

## 11. Editing Existing Workflows: Patterns

### Adding a step mid-workflow
1. Identify the last variable you need — it must be extracted (`extract:`) in an earlier step
2. Insert the new step at the correct position (after its dependencies, before dependent steps)
3. Use `{{ var_name }}` to reference values from earlier `extract:` blocks
4. Add `assert:` to validate the new step's output

### Converting hard-coded values to variables
```yaml
# Before
body:
  email: admin@example.com

# After
body:
  email: "{{ TEST_USER_EMAIL }}"
```
Add to config:
```yaml
config:
  vars:
    TEST_USER_EMAIL: admin@example.com
```

### Debugging a failing workflow
1. Add a `print` step to inspect variables: `message: "Token is {{ token }}"`
2. Use `--debug` flag to see HTTP request/response details
3. Validate first: `replay validate workflow.yaml`
4. Run a single file: `replay run auth/login.yaml`
