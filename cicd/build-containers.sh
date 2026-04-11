#!/bin/bash

docker build -t juno-lan-agent:latest --file Dockerfile.lanagent .
docker build -t juno-conductor --file Dockerfile.conductor .
