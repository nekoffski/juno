#!/usr/bin/env bash
set -euo pipefail

# Starts postgres, builds instrumented binaries, launches them, and waits for
# juno to be ready. Saves background PIDs to coverage/test-pids so that the
# teardown script can stop the processes later.
#
# Requires:
#   - docker compose (for postgres)
#   - go
#   - .env.example present in the repo root

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${REPO_ROOT}/.env.example}"

JUNO_COVER_BIN="${REPO_ROOT}/bin/juno-cover"
JUNO_WEB_COVER_BIN="${REPO_ROOT}/bin/juno-web-cover"

RAW_JUNO="${REPO_ROOT}/coverage/integration-raw/juno"
RAW_WEB="${REPO_ROOT}/coverage/integration-raw/juno-web"
RAW_MERGED="${REPO_ROOT}/coverage/integration-raw/merged"

PID_FILE="${REPO_ROOT}/.test-pids"
NO_COVER="${NO_COVER:-}"

# ------------------------------------------------------------------
# 1. Start postgres
# ------------------------------------------------------------------
echo "--- Starting test environment (postgres) ---"
make -C "${REPO_ROOT}" test-env-up ENV_FILE="${ENV_FILE}"

echo "--- Waiting for postgres to be ready ---"
for i in $(seq 1 30); do
  if ENV_FILE="${ENV_FILE}" docker compose --env-file "${ENV_FILE}" exec -T postgres pg_isready -q 2>/dev/null; then
    echo "postgres is ready (attempt ${i})"
    break
  fi
  if [[ "${i}" -eq 30 ]]; then
    echo "ERROR: postgres did not become ready within 30 seconds" >&2
    exit 1
  fi
  sleep 1
done

# ------------------------------------------------------------------
# 2. Build binaries
# ------------------------------------------------------------------
cd "${REPO_ROOT}"
if [[ "${NO_COVER}" == "1" ]]; then
  echo "--- Building binaries (no coverage) ---"
  go build -o "${JUNO_COVER_BIN}" ./cmd/juno
  go build -o "${JUNO_WEB_COVER_BIN}" ./cmd/juno-web
else
  echo "--- Building instrumented binaries ---"
  go build -cover -o "${JUNO_COVER_BIN}" ./cmd/juno
  go build -cover -o "${JUNO_WEB_COVER_BIN}" ./cmd/juno-web
fi

# ------------------------------------------------------------------
# 3. Prepare raw coverage directories
# ------------------------------------------------------------------
if [[ "${NO_COVER}" != "1" ]]; then
  mkdir -p "${RAW_JUNO}" "${RAW_WEB}" "${RAW_MERGED}"
fi

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
LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "${LOG_DIR}"

echo "--- Starting juno (instrumented) ---"
if [[ "${NO_COVER}" == "1" ]]; then
  "${JUNO_COVER_BIN}" > "${LOG_DIR}/juno.log" 2>&1 &
else
  GOCOVERDIR="${RAW_JUNO}" "${JUNO_COVER_BIN}" > "${LOG_DIR}/juno.log" 2>&1 &
fi
JUNO_PID=$!

echo "--- Starting juno-web (instrumented) ---"
if [[ "${NO_COVER}" == "1" ]]; then
  "${JUNO_WEB_COVER_BIN}" > "${LOG_DIR}/juno-web.log" 2>&1 &
else
  GOCOVERDIR="${RAW_WEB}" "${JUNO_WEB_COVER_BIN}" > "${LOG_DIR}/juno-web.log" 2>&1 &
fi
JUNO_WEB_PID=$!

# Persist PIDs and log dir for the teardown script
printf 'JUNO_PID=%s\nJUNO_WEB_PID=%s\nLOG_DIR=%s\nNO_COVER=%s\n' "${JUNO_PID}" "${JUNO_WEB_PID}" "${LOG_DIR}" "${NO_COVER}" > "${PID_FILE}"

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

echo "--- Setup complete ---"
