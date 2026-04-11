#!/bin/bash

set -euo pipefail

source ./conf/.env.cicd

docker run -d --name postgres --rm \
  --env-file ./conf/.env.cicd \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  -v ./conf/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql:ro \
  --health-cmd "pg_isready -U \$POSTGRES_USER -d \$POSTGRES_DB" \
  --health-interval 5s \
  --health-timeout 5s \
  --health-retries 10 \
  postgres:17-alpine