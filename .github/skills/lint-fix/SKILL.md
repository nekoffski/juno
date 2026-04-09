---
name: lint-fix
description: "Run linters, analyze output, and fix issues. Use when: running linters, fixing lint errors, fixing vet errors, fixing formatting issues, make lint, make vet, make test-fmt, golangci-lint, gofmt, code quality, static analysis."
argument-hint: "Optional: name of a specific linter to run (fmt, vet, lint)"
---

# Lint Fix Workflow

Run all available linters, analyze the output, and propose a prioritized fix plan — then apply fixes if the user approves.

## When to Use

- User asks to "run linters", "fix lint", "check code quality", or "run make lint / make vet"
- Pull request CI is failing on the lint job
- After editing Go files and wanting a quick health check

## Linters in This Project

All linter commands are Makefile targets. Run from the repo root.

| Tool               | Command         | Gate                                    |
| ------------------ | --------------- | --------------------------------------- |
| gofmt (formatting) | `make test-fmt` | CI hard gate — must produce zero output |
| golangci-lint      | `make lint`     | CI soft gate (has known failures)       |
| go vet             | `make vet`      | CI soft gate (has known failure)        |

> `make fmt` auto-formats all files in place. Run it before `make test-fmt` if formatting issues are found.

## Procedure

### Step 1 — Run all linters

Run each linter in order and capture full output:

1. `make test-fmt` — note any files listed (non-zero output = formatting violations)
2. `make lint` — capture full stderr/stdout
3. `make vet` — capture full stderr/stdout

### Step 2 — Analyze and filter output

For each linter result:

- Strip the known pre-existing failures listed above
- Group remaining issues by file
- For each issue note: file path, line number, rule/tool, message

If all remaining issues are zero (only known failures remain), report "No new lint issues found" and stop.

### Step 3 — Plan fixes

Present a numbered fix plan to the user organized by priority:

1. **Formatting** (`make test-fmt` failures) — always auto-fixable via `make fmt`
2. **`go vet` errors** — correctness issues, fix first
3. **`golangci-lint` errors** — style/safety issues

For each item show:

- File and line
- The lint rule and message
- Proposed fix (code snippet or description)

**Ask the user for approval before making any changes.**

### Step 4 — Apply fixes (after user approval)

- For formatting: run `make fmt` — this fixes all formatting automatically
- For other issues: edit the specific files as planned
- After applying all fixes, re-run each relevant linter to confirm zero new issues:
  - `make test-fmt` must produce no output
  - `make build` must succeed
  - `make unit-test` must pass
  - `make functional-tests` must pass

### Step 5 — Report

Summarize what was fixed and the final linter status.
