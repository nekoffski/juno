---
name: code-review
description: "Review code for quality, architecture, and design issues. Use when: code review, design review, architecture review, review git changes, review uncommitted changes, review module, review package, review entire project, PR review, inspect code quality, check design patterns, spot anti-patterns."
argument-hint: "Scope: 'git' (uncommitted changes), 'project' (entire codebase), or a module path like 'internal/device'"
---

# Code Review Workflow

Perform a structured review of code across three scopes — git changes, an entire project, or a specific module — and report findings organized by severity.

## When to Use

- User asks to "review", "code review", "check design", or "review architecture"
- Before merging a PR or committing a batch of changes
- When onboarding to a new package and wanting a quality assessment
- When refactoring and wanting to spot design issues

## Scope Resolution

Determine the review scope from the user's request or argument:

| Scope       | Trigger phrases                                                    | How to gather code                                   |
| ----------- | ------------------------------------------------------------------ | ---------------------------------------------------- |
| **git**     | "git changes", "uncommitted", "what I changed", "my changes"       | `git diff HEAD` + `git diff --cached`                |
| **project** | "entire project", "whole codebase", "everything", no specific path | Scan all packages under `internal/` and `cmd/`       |
| **module**  | A path like `internal/device`, `internal/rest`, `cmd/juno-server`  | List and read all `.go` files in that directory tree |

If the scope is ambiguous, ask before proceeding.

## Procedure

### Step 1 — Gather code for the scope

**git scope:**

```
git diff HEAD
git diff --cached
```

Extract the list of changed `.go` files from the diff. Read the full current version of each changed file (not just the diff lines) to understand context.

**module scope:**

- List all `.go` files under the user-specified path
- Read them in parallel

**project scope:**

- List packages: `internal/`, `cmd/`
- Read key files per package (skip `*.gen.go` — generated, never review directly)
- Also read `api/rest-openapi.yaml` for API design assessment

### Step 2 — Build context

Before reviewing, understand the broader system:

- Read `copilot-instructions.md` for architecture overview (already loaded — use it)
- For module reviews, identify what the module's responsibility is and how it interacts with others
- Note which interfaces the code implements (e.g., `supervisor.Service`, `device.VendorAdapter`, `device.Repository`)

### Step 3 — Review across all dimensions

Evaluate the code on these dimensions. Not every dimension applies to every scope — use judgment.

#### Code Quality

- Idiomatic Go: naming conventions (camelCase vars, PascalCase exports), receiver names, error wrapping
- Error handling: errors propagated or logged correctly, not silently swallowed
- Unnecessary complexity: overly deep nesting, long functions, dead code
- Magic values: unexplained literals that should be constants
- Concurrency: goroutine leaks, missing synchronization, misuse of context

#### Architecture & Design

- Layer violations: does `internal/rest` reach into `internal/db` directly? (should go through device service)
- Dependency direction: dependencies point inward (cmd → internal, never internal → cmd)
- Responsibility: does each package do one thing? Mixed concerns?
- Interface design: interfaces defined at point of use (consumer side), not implementation side
- Service pattern: does each service correctly implement `supervisor.Service` (Name, Init, Run)?
- Message bus usage: are inter-service messages typed correctly, replies always sent?

#### API & Contract Design (REST / MCP)

- OpenAPI spec consistency with handler implementations
- Correct HTTP status codes
- Response shape consistency

#### Security (OWASP Top 10 lens)

- Input validation at system boundaries
- No SQL injection risk (parameterized queries only)
- No sensitive data in logs or error messages
- Auth/authz bypasses

#### Testability

- Pure functions vs. side-effect-heavy code
- Interface seams allowing mocks
- Test coverage gaps on critical paths

### Step 4 — Classify findings

Group all findings by severity:

| Severity       | Meaning                                                                     |
| -------------- | --------------------------------------------------------------------------- |
| **Critical**   | Bug, data corruption risk, security vulnerability, or correctness issue     |
| **Major**      | Design flaw, architecture violation, significant maintainability problem    |
| **Minor**      | Style inconsistency, naming issue, missing error check in non-critical path |
| **Suggestion** | Optional improvement, refactor idea, "nice to have"                         |

### Step 5 — Report

Present findings in this format:

```
## Code Review — <scope>

### Critical
1. [internal/foo/bar.go:42] **Title** — description. Recommendation: ...

### Major
...

### Minor
...

### Suggestions
...

### Summary
N critical, N major, N minor, N suggestions.
Overall assessment: [one sentence on the state of the code]
```

If no issues are found in a severity tier, omit that section.

**Do not make any file edits during the review.** This skill is read-only. If the user wants fixes applied after the report, proceed on their instruction.
