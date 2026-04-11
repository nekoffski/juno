#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

source "${REPO_ROOT}/tests/.venv/bin/activate"
python "${REPO_ROOT}/tests/runner.py" "${REPO_ROOT}/conf/smoke-functional-tests-config.json"
