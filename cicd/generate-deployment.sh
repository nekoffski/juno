#!/bin/bash
set -euo pipefail

DEPLOYMENT_NAME="${1:?Usage: $0 <deployment-name> <env-file> <type (core/edge)>}"
ENV_FILE="${2:?Usage: $0 <deployment-name> <env-file> <type (core/edge)>}"
TYPE="${3:?Usage: $0 <deployment-name> <env-file> <type (core/edge)>}"

if [[ "${TYPE}" != "core" && "${TYPE}" != "edge" ]]; then
    echo "Invalid type: ${TYPE}. Must be 'core' or 'edge'."
    exit 1
fi

set -a
    source "${ENV_FILE}"
set +a

if [[ "${TYPE}" == "core" ]]; then
    export JUNO_LOKI_ADDR="loki"

    DEPLOY_DIR="deployments/core/${DEPLOYMENT_NAME}"
    TEMPLATES=(
        "cicd/template/docker-compose.core.yaml"
        "cicd/template/docker-compose.edge.yaml"
        "cicd/template/prometheus.yaml"
        "cicd/template/loki.yaml"
        "cicd/template/promtail.yaml"
    )

    echo "Creating deployment '${DEPLOYMENT_NAME}' in ${DEPLOY_DIR}"

    if [ -d "${DEPLOY_DIR}" ]; then
        rm -rf "${DEPLOY_DIR}"
    fi

    mkdir -p "${DEPLOY_DIR}"


    for template in "${TEMPLATES[@]}"; do
        output="${DEPLOY_DIR}/$(basename "${template}")"
        echo "  Rendering ${template} -> ${output}"
        envsubst < "${template}" > "${output}"
    done

    cp -rf cicd/template/grafana* "${DEPLOY_DIR}/"
else
    export JUNO_LOKI_ADDR=${JUNO_CORE_ADDR}

    DEPLOY_DIR="deployments/edge/${DEPLOYMENT_NAME}"
    TEMPLATES=(
        "cicd/template/docker-compose.edge.yaml"
    )

    echo "Creating deployment '${DEPLOYMENT_NAME}' in ${DEPLOY_DIR}"

    if [ -d "${DEPLOY_DIR}" ]; then
        rm -rf "${DEPLOY_DIR}"
    fi

    mkdir -p "${DEPLOY_DIR}"

    for template in "${TEMPLATES[@]}"; do
        output="${DEPLOY_DIR}/$(basename "${template}")"
        echo "  Rendering ${template} -> ${output}"
        envsubst < "${template}" > "${output}"
    done
fi


cp -rf "${ENV_FILE}" "${DEPLOY_DIR}/.env"
echo "Done."
