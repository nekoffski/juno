#!/usr/bin/env bash
set -euo pipefail

# Runs the integration test suite against already-running juno + juno-web
# instrumented binaries. Expects integration-test-setup.sh to have been run
# first.
#
# Requires:
#   - tests/.venv already created (make test-venv)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${REPO_ROOT}/.env.example}"

# Load env vars so pytest picks up JUNO_REST_PORT etc.
set -o allexport
# shellcheck source=/dev/null
source <(grep -v '^#' "${ENV_FILE}" | grep '=')
set +o allexport

echo "--- Running integration tests ---"
LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "${LOG_DIR}"
"${REPO_ROOT}/tests/.venv/bin/pytest" "${REPO_ROOT}/tests/" -v \
  --log-file="${LOG_DIR}/pytest.log" \
  --log-file-level=DEBUG \
  --junit-xml="${LOG_DIR}/pytest-results.xml"
