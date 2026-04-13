#!/bin/bash

set -eu 

ENV_TEMPLATE="${1:?Env template missing}"
DEPLOYMENT_NAME="default"
DEPLOYMENT_TYPE=${2:?Deployment type missing (core/edge)}

if [[ "${DEPLOYMENT_TYPE}" != "core" && "${DEPLOYMENT_TYPE}" != "edge" ]]; then
    echo "Invalid deployment type: ${DEPLOYMENT_TYPE}. Must be 'core' or 'edge'."
    exit 1
fi

./cicd/recreate-env.sh ${ENV_TEMPLATE} .env
./cicd/generate-deployment.sh $DEPLOYMENT_NAME .env ${DEPLOYMENT_TYPE}

pushd deployments/${DEPLOYMENT_TYPE}/${DEPLOYMENT_NAME}

docker compose --project-name juno -f docker-compose.${DEPLOYMENT_TYPE}.yaml down -v --remove-orphans
docker compose --project-name juno -f docker-compose.${DEPLOYMENT_TYPE}.yaml pull
docker compose --project-name juno -f docker-compose.${DEPLOYMENT_TYPE}.yaml up -d

popd 
