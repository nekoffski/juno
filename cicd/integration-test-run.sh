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

echo "--- Running integration tests ---"
"${REPO_ROOT}/tests/.venv/bin/pytest" "${REPO_ROOT}/tests/" -v
