#!/bin/bash

docker build -t juno-lan-agent:latest .

ENV_FILE=./conf/.env.example.docker docker compose --project-name juno \
  --env-file ./conf/.env.example.docker -f docker-compose.yaml build
