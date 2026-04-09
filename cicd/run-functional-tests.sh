#!/bin/bash
set -euo pipefail

source ./tests/.venv/bin/activate
python ./tests/runner.py ./conf/full-functional-tests-config.json
