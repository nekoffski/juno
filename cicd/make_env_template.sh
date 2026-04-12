#!/bin/bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <env-file>" >&2
    exit 1
fi

ENV_FILE="$1"

if [[ ! -f "$ENV_FILE" ]]; then
    echo "Error: file '$ENV_FILE' not found" >&2
    exit 1
fi

grep -v '^\s*#' "$ENV_FILE" | grep -v '^\s*$' | awk 'NR>1{printf "&&"} {printf "%s", $0} END{printf "\n"}'