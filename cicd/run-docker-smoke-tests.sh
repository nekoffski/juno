#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

source "${REPO_ROOT}/tests/.venv/bin/activate"
python "${REPO_ROOT}/tests/runner.py" "${REPO_ROOT}/conf/smoke-functional-tests-config.json"
