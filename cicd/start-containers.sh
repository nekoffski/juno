#!/bin/bash

docker run -d --name juno-lan-agent --rm \
    --network host --env-file ./conf/.env.example.docker juno-lan-agent:latest

ENV_FILE=./conf/.env.example.docker docker compose --project-name juno \
  --env-file ./conf/.env.example.docker -f docker-compose.yaml up -d
