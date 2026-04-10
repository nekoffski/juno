#!/bin/bash

docker run -d --name juno-lan-agent --rm \
    --network host --env-file ./conf/.env.example.docker juno-lan-agent:latest
