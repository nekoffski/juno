#!/usr/bin/env bash
set -euo pipefail

# Gracefully stops juno-conductor (which forwards SIGTERM to all child
# processes, flushing coverage data), generates HTML + Cobertura XML reports,
# and tears down the postgres container.
#
# Must be run after integration-test-setup.sh. Reads the conductor PID from
# .test-pids written by the setup script.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${REPO_ROOT}/conf/.env.example}"

RAW_ALL="${REPO_ROOT}/coverage/integration-raw/all"
CONDUCTOR_TEST_CFG="${REPO_ROOT}/.conductor-test.yaml"
INTEGRATION_PROFILE="${REPO_ROOT}/integration-coverage.txt"
REPORT_DIR="${REPO_ROOT}/coverage/integration"

PID_FILE="${REPO_ROOT}/.test-pids"

# ------------------------------------------------------------------
# 1. Stop conductor (forwards SIGTERM to children, flushing coverage)
# ------------------------------------------------------------------
echo "--- Stopping conductor ---"
if [[ -f "${PID_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${PID_FILE}"
  if [[ -n "${CONDUCTOR_PID:-}" ]] && kill -0 "${CONDUCTOR_PID}" 2>/dev/null; then
    echo "Stopping conductor (pid ${CONDUCTOR_PID})"
    kill -TERM "${CONDUCTOR_PID}" || true
    for _ in $(seq 1 60); do
      kill -0 "${CONDUCTOR_PID}" 2>/dev/null || break
      sleep 0.1
    done
  fi
  if [[ -n "${LAN_AGENT_PID:-}" ]] && kill -0 "${LAN_AGENT_PID}" 2>/dev/null; then
    echo "Stopping lan-agent (pid ${LAN_AGENT_PID})"
    kill -TERM "${LAN_AGENT_PID}" || true
  fi
  rm -f "${PID_FILE}"
else
  echo "WARNING: PID file not found at ${PID_FILE}, conductor may still be running" >&2
fi
rm -f "${CONDUCTOR_TEST_CFG}"

# ------------------------------------------------------------------
# 2. Merge raw coverage from both binaries
# ------------------------------------------------------------------
if [[ "${NO_COVER:-}" == "1" ]]; then
  echo "--- Skipping coverage (NO_COVER is set) ---"
else
  # ------------------------------------------------------------------
  # 2. Convert raw coverage to text profile format
  # ------------------------------------------------------------------
  echo "--- Converting coverage to profile format ---"
  go tool covdata textfmt -i="${RAW_ALL}" -o="${INTEGRATION_PROFILE}"

  # ------------------------------------------------------------------
  # 3. Generate HTML + XML reports
  # ------------------------------------------------------------------
  echo "--- Generating coverage reports ---"
  bash "${SCRIPT_DIR}/generate-coverage-reports.sh" "${INTEGRATION_PROFILE}" "${REPORT_DIR}"
fi

# ------------------------------------------------------------------
# 5. Stop postgres
# ------------------------------------------------------------------
echo "--- Stopping test environment ---"
make -C "${REPO_ROOT}" test-env-down ENV_FILE="${ENV_FILE}"

echo "--- Teardown complete ---"
if [[ "${NO_COVER:-}" != "1" ]]; then
  echo "Reports available in: ${REPORT_DIR}"
fi
