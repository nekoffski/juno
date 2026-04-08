#!/usr/bin/env bash
set -euo pipefail

# Starts postgres, builds instrumented binaries, launches them via juno-conductor,
# and waits for juno-server to be ready. Saves the conductor PID to .test-pids
# so that the teardown script can stop all processes later.
#
# Requires:
#   - docker compose (for postgres)
#   - go
#   - conf/.env.example present in the repo root

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${REPO_ROOT}/conf/.env.example}"

JUNO_SERVER_BIN="${REPO_ROOT}/bin/juno-server-cover"
JUNO_WEB_BIN="${REPO_ROOT}/bin/juno-web-cover"
JUNO_MCP_BIN="${REPO_ROOT}/bin/juno-mcp-cover"
JUNO_CONDUCTOR_BIN="${REPO_ROOT}/bin/juno-conductor-cover"
JUNO_LAN_BIN="${REPO_ROOT}/bin/juno-lan-agent-cover"

RAW_ALL="${REPO_ROOT}/coverage/integration-raw/all"
CONDUCTOR_TEST_CFG="${REPO_ROOT}/.conductor-test.yaml"

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
echo "--- Building binaries ---"
if [[ "${NO_COVER}" == "1" ]]; then
  go build -o "${JUNO_SERVER_BIN}"    ./cmd/juno-server
  go build -o "${JUNO_WEB_BIN}"       ./cmd/juno-web
  go build -o "${JUNO_MCP_BIN}"       ./cmd/juno-mcp
  go build -o "${JUNO_CONDUCTOR_BIN}" ./cmd/juno-conductor
  go build -o "${JUNO_LAN_BIN}"       ./cmd/juno-lan-agent
else
  go build -cover -o "${JUNO_SERVER_BIN}"    ./cmd/juno-server
  go build -cover -o "${JUNO_WEB_BIN}"       ./cmd/juno-web
  go build -cover -o "${JUNO_MCP_BIN}"       ./cmd/juno-mcp
  go build -cover -o "${JUNO_CONDUCTOR_BIN}" ./cmd/juno-conductor
  go build -cover -o "${JUNO_LAN_BIN}"       ./cmd/juno-lan-agent
fi

# ------------------------------------------------------------------
# 3. Prepare raw coverage directories
# ------------------------------------------------------------------
mkdir -p "${RAW_ALL}"

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
# Use overrides for local testing: agent runs on localhost,
# JUNO_LAN_AGENT_PORT/ADDR come from ENV_FILE (default 7000 / 0.0.0.0)
JUNO_LAN_AGENT_PORT="${JUNO_LAN_AGENT_PORT:-7000}"
JUNO_LAN_AGENT_ADDR="${JUNO_LAN_AGENT_ADDR:-127.0.0.1}"
JUNO_LAN_AGENT_URL="http://localhost:${JUNO_LAN_AGENT_PORT}"
export JUNO_LAN_AGENT_URL JUNO_LAN_AGENT_PORT JUNO_LAN_AGENT_ADDR

# ------------------------------------------------------------------
# 4a. Generate conductor test config pointing at instrumented binaries
# ------------------------------------------------------------------
cat > "${CONDUCTOR_TEST_CFG}" <<EOF
processes:
  - name: juno-server
    binary: ${JUNO_SERVER_BIN}
  - name: juno-web
    binary: ${JUNO_WEB_BIN}
  - name: juno-mcp
    binary: ${JUNO_MCP_BIN}
EOF

# ------------------------------------------------------------------
# 5. Start lan-agent
# ------------------------------------------------------------------
LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "${LOG_DIR}"

echo "--- Starting lan-agent on ${JUNO_LAN_AGENT_ADDR}:${JUNO_LAN_AGENT_PORT} ---"
if [[ "${NO_COVER}" == "1" ]]; then
  JUNO_LAN_AGENT_PORT="${JUNO_LAN_AGENT_PORT}" JUNO_LAN_AGENT_ADDR="${JUNO_LAN_AGENT_ADDR}" "${JUNO_LAN_BIN}" > "${LOG_DIR}/lan-agent.log" 2>&1 &
else
  GOCOVERDIR="${RAW_ALL}" JUNO_LAN_AGENT_PORT="${JUNO_LAN_AGENT_PORT}" JUNO_LAN_AGENT_ADDR="${JUNO_LAN_AGENT_ADDR}" "${JUNO_LAN_BIN}" > "${LOG_DIR}/lan-agent.log" 2>&1 &
fi
LAN_AGENT_PID=$!

for i in $(seq 1 15); do
  if curl -sf "http://localhost:${JUNO_LAN_AGENT_PORT}/health" > /dev/null 2>&1; then
    echo "lan-agent is ready (attempt ${i})"
    break
  fi
  if [[ "${i}" -eq 15 ]]; then
    echo "ERROR: lan-agent did not become ready" >&2
    exit 1
  fi
  sleep 0.5
done

# ------------------------------------------------------------------
# 5b. Start conductor (manages all instrumented binaries)
# ------------------------------------------------------------------

echo "--- Starting conductor ---"
if [[ "${NO_COVER}" == "1" ]]; then
  "${JUNO_CONDUCTOR_BIN}" -config "${CONDUCTOR_TEST_CFG}" > "${LOG_DIR}/conductor.log" 2>&1 &
else
  GOCOVERDIR="${RAW_ALL}" "${JUNO_CONDUCTOR_BIN}" -config "${CONDUCTOR_TEST_CFG}" > "${LOG_DIR}/conductor.log" 2>&1 &
fi
CONDUCTOR_PID=$!

# Persist PIDs and log dir for the teardown script
printf 'CONDUCTOR_PID=%s\nLAN_AGENT_PID=%s\nLOG_DIR=%s\nNO_COVER=%s\n' "${CONDUCTOR_PID}" "${LAN_AGENT_PID}" "${LOG_DIR}" "${NO_COVER}" > "${PID_FILE}"

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
