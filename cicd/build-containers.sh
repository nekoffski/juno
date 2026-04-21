#!/bin/bash

docker buildx build --platform linux/amd64,linux/arm/v6 --push -t juno-lan-agent:latest --file Dockerfile.lanagent .
docker build -t juno-conductor --file Dockerfile.conductor .
