#!/bin/bash

set -eu

VENV=cicd/deployer/.venv

if ! [ -d  "$VENV" ]; then
    python3 -m venv "$VENV"
fi
source "$VENV/bin/activate"
pip install -r cicd/deployer/requirements.txt > /dev/null

python cicd/deployer/deployer.py $@
