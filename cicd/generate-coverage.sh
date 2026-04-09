#!/usr/bin/env bash
set -euo pipefail

go install github.com/boumenot/gocover-cobertura@latest
go tool covdata textfmt -i="./coverage/raw" -o="./coverage/functional-coverage.txt"
go tool cover -html="./coverage/functional-coverage.txt" -o "./coverage/index.html"
