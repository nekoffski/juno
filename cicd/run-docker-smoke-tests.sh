#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${REPO_ROOT}/conf/.env.example.docker"
COMPOSE_PROJECT="juno"
LAN_AGENT_CONTAINER="juno-lan-agent"
LAN_AGENT_IMAGE="juno-lan-agent"

cleanup() {
  docker stop "${LAN_AGENT_CONTAINER}" 2>/dev/null || true
  docker rm   "${LAN_AGENT_CONTAINER}" 2>/dev/null || true

  ENV_FILE="${ENV_FILE}" docker compose --project-name "${COMPOSE_PROJECT}" \
    --env-file "${ENV_FILE}" -f "${REPO_ROOT}/docker-compose.yaml" down --volumes
}
trap cleanup EXIT

ENV_FILE="${ENV_FILE}" docker compose --project-name "${COMPOSE_PROJECT}" \
  --env-file "${ENV_FILE}" -f "${REPO_ROOT}/docker-compose.yaml" up -d

docker run -d \
  --name "${LAN_AGENT_CONTAINER}" \
  --network host \
  --env-file "${ENV_FILE}" \
  "${LAN_AGENT_IMAGE}"

source <(grep -v '^#' "${ENV_FILE}" | grep '=')
JUNO_REST_PORT="${JUNO_REST_PORT:-6001}"
HEALTH_URL="http://localhost:${JUNO_REST_PORT}/health"

for i in $(seq 1 30); do
  if curl -sf "${HEALTH_URL}" > /dev/null 2>&1; then
    echo "juno is ready (attempt ${i})"
    break
  fi
  if [[ "${i}" -eq 30 ]]; then
    echo "ERROR: juno did not become ready" >&2
    exit 1
  fi
  sleep 1
done

source "${REPO_ROOT}/tests/.venv/bin/activate"
python "${REPO_ROOT}/tests/runner.py" "${REPO_ROOT}/conf/smoke-functional-tests-config.json"
