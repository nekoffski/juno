#!/usr/bin/env bash
# Run the full integration test suite locally using .env.example.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

EXIT_CODE=0

NO_COVER=1 ENV_FILE="${REPO_ROOT}/.env.example" bash "${SCRIPT_DIR}/integration-test-setup.sh"
NO_COVER=1 ENV_FILE="${REPO_ROOT}/.env.example" bash "${SCRIPT_DIR}/integration-test-run.sh" || EXIT_CODE=$?
NO_COVER=1 ENV_FILE="${REPO_ROOT}/.env.example" bash "${SCRIPT_DIR}/integration-test-teardown.sh"

if [[ "${EXIT_CODE}" -ne 0 ]]; then
  echo ""
  echo "--- Tests FAILED (exit ${EXIT_CODE}) ---"
  echo "Logs are in ${REPO_ROOT}/logs/"
fi

exit "${EXIT_CODE}"
