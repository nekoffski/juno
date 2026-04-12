#!/bin/bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
    echo "Usage: $0 <env-string> <output-file>" >&2
    exit 1
fi

ENV_STRING="$1"
OUTPUT_FILE="$2"

echo "$ENV_STRING" | sed 's/&&/\n/g' > "$OUTPUT_FILE"
