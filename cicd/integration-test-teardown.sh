#!/usr/bin/env bash
set -euo pipefail

# Gracefully stops instrumented juno + juno-web binaries (flushing coverage
# data), merges the raw coverage, generates HTML + Cobertura XML reports, and
# tears down the postgres container.
#
# Must be run after integration-test-setup.sh. Reads PIDs from
# coverage/test-pids written by the setup script.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${REPO_ROOT}/.env.example}"

RAW_JUNO="${REPO_ROOT}/coverage/integration-raw/juno"
RAW_WEB="${REPO_ROOT}/coverage/integration-raw/juno-web"
RAW_MERGED="${REPO_ROOT}/coverage/integration-raw/merged"
INTEGRATION_PROFILE="${REPO_ROOT}/8"
REPORT_DIR="${REPO_ROOT}/coverage/integration"

PID_FILE="${REPO_ROOT}/.test-pids"

# ------------------------------------------------------------------
# 1. Stop instrumented binaries (flushes coverage data)
# ------------------------------------------------------------------
echo "--- Stopping instrumented binaries ---"
if [[ -f "${PID_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${PID_FILE}"
  if [[ -n "${JUNO_PID:-}" ]] && kill -0 "${JUNO_PID}" 2>/dev/null; then
    echo "Stopping juno (pid ${JUNO_PID})"
    kill -TERM "${JUNO_PID}" || true
    # wait is only valid for child processes; poll instead
    for _ in $(seq 1 30); do
      kill -0 "${JUNO_PID}" 2>/dev/null || break
      sleep 0.1
    done
  fi
  if [[ -n "${JUNO_WEB_PID:-}" ]] && kill -0 "${JUNO_WEB_PID}" 2>/dev/null; then
    echo "Stopping juno-web (pid ${JUNO_WEB_PID})"
    kill -TERM "${JUNO_WEB_PID}" || true
    for _ in $(seq 1 30); do
      kill -0 "${JUNO_WEB_PID}" 2>/dev/null || break
      sleep 0.1
    done
  fi
  rm -f "${PID_FILE}"
else
  echo "WARNING: PID file not found at ${PID_FILE}, binaries may still be running" >&2
fi

# ------------------------------------------------------------------
# 2. Merge raw coverage from both binaries
# ------------------------------------------------------------------
if [[ "${NO_COVER:-}" == "1" ]]; then
  echo "--- Skipping coverage (NO_COVER is set) ---"
else
  echo "--- Merging coverage data ---"
  go tool covdata merge -i="${RAW_JUNO},${RAW_WEB}" -o="${RAW_MERGED}"

  # ------------------------------------------------------------------
  # 3. Convert merged raw data to text profile format
  # ------------------------------------------------------------------
  echo "--- Converting merged coverage to profile format ---"
  go tool covdata textfmt -i="${RAW_MERGED}" -o="${INTEGRATION_PROFILE}"

  # ------------------------------------------------------------------
  # 4. Generate HTML + XML reports
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
