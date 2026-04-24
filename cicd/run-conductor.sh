#!/bin/bash

set -euo pipefail

CONFIG_FILE="${1:-./conf/conductor.local.json}"

set -a
source ./conf/.env.example
set +a

exec ./bin/juno-conductor -config "$CONFIG_FILE"
