#!/bin/bash

set -eu

VENV=cicd/deployer/.venv

if ! [ -d  "$VENV" ]; then
    python3 -m venv "$VENV" > /dev/null 2>&1
fi
source "$VENV/bin/activate"
pip install -r cicd/deployer/requirements.txt > /dev/null 2>&1

python cicd/deployer/deployer.py $@ > /dev/null 2>&1
