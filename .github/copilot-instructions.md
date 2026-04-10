# Juno — Copilot Agent Instructions

Trust these instructions. Only search the codebase if the information here is incomplete or appears incorrect.

> **Maintainers:** Keep this file up to date as the codebase evolves — especially when binaries are added/removed, the service design changes, or new frameworks are introduced.

## What This Repository Does

Juno is a home automation server written in Go. It exposes a REST API, a web UI, and an MCP (Model Context Protocol) server for device control. Devices are discovered/controlled via vendor adapters (currently Yeelight).

### Deployment topology

The system is split into two independent deployment units:

**Core services container** — Built by `Dockerfile`, managed by `docker-compose.yaml`. Contains three binaries supervised by `juno-conductor`:

- `juno-server` — REST API + device service + PostgreSQL access. Multiple internal sub-services (device, REST) wired together through the `supervisor.Supervisor` and the in-process message bus.
- `juno-web` — Web UI, proxies requests to `juno-server`.
- `juno-mcp` — MCP server, proxies requests to `juno-server`.

`juno-conductor` is a process supervisor that starts, restarts-on-crash, and shuts down these three binaries. It is not a Go `supervisor.Service`; it operates at the OS-process level.

PostgreSQL runs as a separate service in the same `docker-compose.yaml`.

**LAN agent container** — Built by `Dockerfile.lan`, deployed as a separate container **inside the local network**. Contains only `juno-lan-agent`. This binary is intentionally simple: it has no supervisor, no message bus, and no database. It exposes a plain HTTP API (`/health`, `/discover`, CONNECT proxy) that `juno-server` calls to reach LAN devices (e.g. Yeelight bulbs). Do not apply the `supervisor.Service` pattern to this binary — it does not need it.

## Language, Runtime, and Tools

- **Language**: Go 1.26.1 (module: `github.com/nekoffski/juno`)
- **Key frameworks**: Echo v4 (REST), pgx v5 (PostgreSQL), golang-migrate, oapi-codegen, MCP Go SDK
- **Database**: PostgreSQL 17; migrations are embedded in the binary via `//go:embed`
- **Config**: environment variables parsed with `caarlos0/env`; local defaults in `conf/.env.example`
- **REST code gen**: `oapi-codegen` v1.16.3 (CI-pinned version); spec lives at `api/rest-openapi.yaml`
- **Linter**: `golangci-lint` v1.64.8 (CI-pinned)
- **Functional tests**: Python 3 + pytest, in `tests/`

## Build & Validate — Always Use Make Targets

All commands must be run from the repository root.

### One-time tool install (required before first build)

```
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.16.3
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

### Normal development workflow (in order)

| Purpose                                 | Command                |
| --------------------------------------- | ---------------------- |
| Tidy dependencies                       | `make tidy`            |
| Generate REST API code                  | `make generate`        |
| Build all binaries → `bin/`             | `make build`           |
| Run unit tests                          | `make unit-test`       |
| Run unit tests with coverage            | `make unit-test-cover` |
| Format code                             | `make fmt`             |
| Check formatting (CI gate)              | `make test-fmt`        |
| Run linter                              | `make lint`            |
| Run vet                                 | `make vet`             |
| Clean build artifacts & generated files | `make clean`           |

**Always run `make generate` before `make build`** — `make build` depends on `generate` and will fail without the generated file `internal/rest/api.gen.go`.

After `make clean`, generated files (`*.gen.go`) are deleted; `make build` will regenerate them automatically (it depends on `make generate`).

### Reliable validation gates for new code

1. `make test-fmt` — must produce no output (zero exit code)
2. `make build` — must succeed
3. `make unit-test` — all packages must pass
4. `make functional-tests` — all tests must pass (requires test venv setup, locally it should be already generated)

## Architecture and Key File Locations

```
cmd/                        # main() entry points — one directory per binary (list may grow/shrink)
  juno-server/main.go       # REST API + device service + DB
  juno-web/main.go          # Web UI service
  juno-mcp/main.go          # MCP server
  juno-conductor/main.go    # Process supervisor (manages other binaries)
  juno-lan-agent/main.go    # LAN agent — standalone binary, separate LAN-network container
                            #   Simple HTTP server only; no supervisor, no bus, no DB.
                            #   NOTE: the canonical binary list lives in TARGETS in the Makefile

internal/
  core/config.go            # Env-var config structs (LoadConfig)
  db/db.go                  # PostgreSQL pool, runs embedded migrations on Open()
  db/migrations/            # SQL migration files (embedded at compile time)
  device/                   # Device model, repository interface, service, vendor adapter interface
  yeelight/                 # Yeelight vendor adapter (implements device.VendorAdapter)
  rest/                     # Echo REST handlers; api.gen.go is GENERATED — never edit directly
  web/                      # Web service
  mcp/                      # MCP server service
  lan/                      # LAN agent HTTP handlers
  supervisor/               # Service lifecycle management (Init → Run pattern)
  bus/                      # In-process message bus

api/
  rest-openapi.yaml         # OpenAPI 3.0 spec — source of truth for REST API
  rest-oapi-codegen.yaml    # oapi-codegen config: outputs to internal/rest/api.gen.go

conf/
  .env.example              # Local dev environment variables (copy and fill in)
  .env.example.docker       # Docker env vars
  conductor.json            # Conductor config for containerized deployment
  conductor.local.json      # Conductor config for local functional tests
  postgres/init.sql         # Postgres DB initialization

.github/workflows/test.yaml # CI pipeline (lint → build+unit-test → functional → docker)
Makefile                    # All scripted steps
Dockerfile                  # Multi-stage build for core services (juno-server, juno-web, juno-mcp, juno-conductor)
Dockerfile.lan              # Build for juno-lan-agent (separate container, deployed in LAN)
docker-compose.yaml         # PostgreSQL + conductor container (runs core services)
```

## Adding or Changing REST API Endpoints

If editing the existing server:

1. Edit `api/rest-openapi.yaml`
2. Run `make generate` — this regenerates `internal/rest/api.gen.go`
3. Implement the new handler method in `internal/rest/handlers_*.go`
4. Never edit `internal/rest/api.gen.go` directly; it will be overwritten

If adding new binaries/services, follow the existing patterns in `cmd/` and `internal/` for structure and organization.

## CI Pipeline (`.github/workflows/test.yaml`)

Runs on every push and PR. Three jobs run in sequence:

1. **lint**: `make test-fmt` → `make lint` → `make vet`
2. **build-and-unit-test**: `make generate` → `make build` → `make unit-test-cover`
3. **functional-tests** (needs job 2): `make generate` → `make coverage-build` → `make test-venv` → `./cicd/run-functional-tests.sh`
4. **docker-build** (needs job 2): builds Docker images + smoke tests

Functional tests require a running PostgreSQL instance (started by the test runner via Docker) and binaries compiled with `-cover` (`make coverage-build`).

## Service Design Pattern

**Applies to `juno-server` only.** The core server binary wires multiple internal services through `supervisor.Supervisor`, which calls `Init` then `Run` concurrently on each. The message bus (`internal/bus`) is passed to each service during `Init` for decoupled inter-service communication.

Services following this pattern: `device.DeviceService`, `rest.RestService`.

**Does NOT apply to `juno-lan-agent`.** The LAN agent is a standalone HTTP server (`lan.Service`) with a simple `Run(ctx)` method and no supervisor, no bus, and no DB. When adding code to `internal/lan`, keep it simple — plain `net/http`, context cancellation for shutdown, nothing more.

**Does NOT apply to `juno-web` or `juno-mcp`.** These binaries are standalone HTTP servers that forward requests to `juno-server` over HTTP. They have no internal bus and do not implement `supervisor.Service`.

## Environment Variables (key ones)

| Var                  | Default    | Purpose              |
| -------------------- | ---------- | -------------------- |
| `POSTGRES_USER`      | required   | DB user              |
| `POSTGRES_PASSWORD`  | required   | DB password          |
| `POSTGRES_DB`        | required   | DB name              |
| `POSTGRES_HOST`      | `postgres` | DB host              |
| `JUNO_REST_PORT`     | `6000`     | REST API port        |
| `JUNO_WEB_PORT`      | `6001`     | Web UI port          |
| `JUNO_LAN_AGENT_URL` | ``         | URL of the LAN agent |

## Code Style

- Use `make fmt` to auto-format code; this is the canonical style (enforced by CI)
- For Go code, follow standard Go conventions (e.g. camelCase for variables, PascalCase for exported identifiers)
- For REST API handlers, follow the patterns in `internal/rest/handlers_*.go` (e.g. separate files by domain, use Echo context, return proper HTTP status codes)
- For database access, use the repository pattern as in `internal/device/repository.go` and follow the existing style for SQL queries and error handling
- For new services, follow the `supervisor.Service` pattern and use the message bus for communication with other services when needed
- Do not add comments that state the obvious; code should be self-explanatory. Add comments for non-trivial logic or decisions. Functional tests are exception, the test case itself can be explained with 2/3 sentences of comments at the top of the test function, but individual assertions should not be commented.
