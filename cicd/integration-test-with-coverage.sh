#!/usr/bin/env bash
set -euo pipefail

# Runs integration tests against instrumented juno + juno-web binaries and
# generates merged Go coverage reports (HTML + Cobertura XML).
#
# Requires:
#   - docker compose (for postgres)
#   - go, gocover-cobertura
#   - tests/.venv already created (make test-venv)
#   - .env.example present in the repo root

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${REPO_ROOT}/.env.example}"

JUNO_COVER_BIN="${REPO_ROOT}/bin/juno-cover"
JUNO_WEB_COVER_BIN="${REPO_ROOT}/bin/juno-web-cover"

RAW_JUNO="${REPO_ROOT}/coverage/integration-raw/juno"
RAW_WEB="${REPO_ROOT}/coverage/integration-raw/juno-web"
RAW_MERGED="${REPO_ROOT}/coverage/integration-raw/merged"
INTEGRATION_PROFILE="${REPO_ROOT}/integration-coverage.out"
REPORT_DIR="${REPO_ROOT}/coverage/integration"

JUNO_PID=""
JUNO_WEB_PID=""

cleanup() {
  echo "--- Cleanup ---"
  if [[ -n "${JUNO_PID}" ]] && kill -0 "${JUNO_PID}" 2>/dev/null; then
    echo "Stopping juno (pid ${JUNO_PID})"
    kill -TERM "${JUNO_PID}" && wait "${JUNO_PID}" || true
  fi
  if [[ -n "${JUNO_WEB_PID}" ]] && kill -0 "${JUNO_WEB_PID}" 2>/dev/null; then
    echo "Stopping juno-web (pid ${JUNO_WEB_PID})"
    kill -TERM "${JUNO_WEB_PID}" && wait "${JUNO_WEB_PID}" || true
  fi
  echo "Stopping test environment"
  make -C "${REPO_ROOT}" test-env-down ENV_FILE="${ENV_FILE}" || true
}
trap cleanup EXIT

# ------------------------------------------------------------------
# 1. Start postgres
# ------------------------------------------------------------------
echo "--- Starting test environment (postgres) ---"
make -C "${REPO_ROOT}" test-env-up ENV_FILE="${ENV_FILE}"

# ------------------------------------------------------------------
# 2. Build instrumented binaries
# ------------------------------------------------------------------
echo "--- Building instrumented binaries ---"
cd "${REPO_ROOT}"
go build -cover -o "${JUNO_COVER_BIN}" ./cmd/juno
go build -cover -o "${JUNO_WEB_COVER_BIN}" ./cmd/juno-web

# ------------------------------------------------------------------
# 3. Prepare raw coverage directories
# ------------------------------------------------------------------
mkdir -p "${RAW_JUNO}" "${RAW_WEB}" "${RAW_MERGED}"

# ------------------------------------------------------------------
# 4. Load env vars from ENV_FILE (skip comments and blank lines)
# ------------------------------------------------------------------
set -o allexport
# shellcheck source=/dev/null
source <(grep -v '^#' "${ENV_FILE}" | grep '=')
set +o allexport

# Postgres is exposed on localhost when binaries run directly on the host
# (POSTGRES_HOST=postgres is the Docker service name, only valid inside Docker)
POSTGRES_HOST=localhost
# juno-web proxies to the REST API, also on localhost when running on the host
JUNO_REST_BASE_URL="http://localhost:${JUNO_REST_PORT:-6001}"

# ------------------------------------------------------------------
# 5. Start instrumented binaries in background
# ------------------------------------------------------------------
echo "--- Starting juno (instrumented) ---"
GOCOVERDIR="${RAW_JUNO}" "${JUNO_COVER_BIN}" &
JUNO_PID=$!

echo "--- Starting juno-web (instrumented) ---"
GOCOVERDIR="${RAW_WEB}" "${JUNO_WEB_COVER_BIN}" &
JUNO_WEB_PID=$!

# ------------------------------------------------------------------
# 6. Wait for juno REST to be ready
# ------------------------------------------------------------------
JUNO_REST_PORT="${JUNO_REST_PORT:-6001}"
HEALTH_URL="http://localhost:${JUNO_REST_PORT}/health"
echo "--- Waiting for juno to be ready at ${HEALTH_URL} ---"
for i in $(seq 1 30); do
  if curl -sf "${HEALTH_URL}" > /dev/null 2>&1; then
    echo "juno is ready (attempt ${i})"
    break
  fi
  if [[ "${i}" -eq 30 ]]; then
    echo "ERROR: juno did not become ready within 30 seconds" >&2
    exit 1
  fi
  sleep 1
done

# ------------------------------------------------------------------
# 7. Run integration tests
# ------------------------------------------------------------------
echo "--- Running integration tests ---"
"${REPO_ROOT}/tests/.venv/bin/pytest" "${REPO_ROOT}/tests/" -v

# ------------------------------------------------------------------
# 8. Gracefully stop instrumented binaries (flushes coverage)
# ------------------------------------------------------------------
echo "--- Stopping instrumented binaries ---"
kill -TERM "${JUNO_PID}" && wait "${JUNO_PID}" || true; JUNO_PID=""
kill -TERM "${JUNO_WEB_PID}" && wait "${JUNO_WEB_PID}" || true; JUNO_WEB_PID=""

# ------------------------------------------------------------------
# 9. Merge raw coverage from both binaries
# ------------------------------------------------------------------
echo "--- Merging coverage data ---"
go tool covdata merge -i="${RAW_JUNO},${RAW_WEB}" -o="${RAW_MERGED}"

# ------------------------------------------------------------------
# 10. Convert merged raw data to text profile format
# ------------------------------------------------------------------
echo "--- Converting merged coverage to profile format ---"
go tool covdata textfmt -i="${RAW_MERGED}" -o="${INTEGRATION_PROFILE}"

# ------------------------------------------------------------------
# 11. Generate HTML + XML reports
# ------------------------------------------------------------------
echo "--- Generating coverage reports ---"
bash "${SCRIPT_DIR}/generate-coverage-reports.sh" "${INTEGRATION_PROFILE}" "${REPORT_DIR}"

echo "--- Integration test coverage complete ---"
echo "Reports available in: ${REPORT_DIR}"
