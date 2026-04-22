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
    export JUNO_LOKI_URL="loki:3100"

    DEPLOY_DIR="deployments/core/${DEPLOYMENT_NAME}"
    TEMPLATES=(
        "cicd/template/docker-compose.core.yaml"
        "cicd/template/prometheus.yaml"
        "cicd/template/loki.yaml"
        "cicd/template/promtail.yaml"
        "cicd/template/nginx.conf"
    )

    echo "Creating deployment '${DEPLOYMENT_NAME}' in ${DEPLOY_DIR}"
    mkdir -p "${DEPLOY_DIR}"


    SUBST_VARS=$(grep -E '^[A-Z_]+=' "${ENV_FILE}" | cut -d= -f1 | sed 's/^/\${/' | sed 's/$/}/' | tr '\n' ' ')
    SUBST_VARS="${SUBST_VARS} \${JUNO_LOKI_URL}"

    for template in "${TEMPLATES[@]}"; do
        output="${DEPLOY_DIR}/$(basename "${template}")"
        echo "  Rendering ${template} -> ${output}"
        envsubst "${SUBST_VARS}" < "${template}" > "${output}"
    done

    cp conf/postgres/init.sql "${DEPLOY_DIR}/init.sql"
    cp -rf cicd/template/grafana* "${DEPLOY_DIR}/"
else
    export JUNO_LOKI_URL="${JUNO_CORE_ADDR}:${JUNO_NGINX_PORT}/_loki"

    DEPLOY_DIR="deployments/edge/${DEPLOYMENT_NAME}"
    TEMPLATES=(
        "cicd/template/docker-compose.edge.yaml"
        "cicd/template/promtail.yaml"
    )

    echo "Creating deployment '${DEPLOYMENT_NAME}' in ${DEPLOY_DIR}"
    mkdir -p "${DEPLOY_DIR}"

    for template in "${TEMPLATES[@]}"; do
        output="${DEPLOY_DIR}/$(basename "${template}")"
        echo "  Rendering ${template} -> ${output}"
        envsubst < "${template}" > "${output}"
    done
fi


cp -rf "${ENV_FILE}" "${DEPLOY_DIR}/.env"
echo "Done."
