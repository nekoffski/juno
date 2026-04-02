#!/usr/bin/env bash
set -euo pipefail

# Usage: generate-coverage-reports.sh <coverage.out> <output-dir>
#
# Generates HTML and Cobertura XML coverage reports from a Go coverage profile.
# Requires: go, gocover-cobertura (go install github.com/boumenot/gocover-cobertura@latest)

COVERAGE_OUT="${1:?Usage: $0 <coverage.out> <output-dir>}"
OUTPUT_DIR="${2:?Usage: $0 <coverage.out> <output-dir>}"

mkdir -p "${OUTPUT_DIR}"

echo "Generating HTML report -> ${OUTPUT_DIR}/index.html"
go tool cover -html="${COVERAGE_OUT}" -o "${OUTPUT_DIR}/index.html"

echo "Generating Cobertura XML report -> ${OUTPUT_DIR}/coverage.xml"
gocover-cobertura < "${COVERAGE_OUT}" > "${OUTPUT_DIR}/coverage.xml"

echo "Coverage reports written to ${OUTPUT_DIR}"
