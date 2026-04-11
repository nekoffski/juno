#!/bin/bash

echo $DOCKER_TOKEN | docker login --username $DOCKER_USER --password-stdin


docker tag juno-lan-agent $DOCKER_USER/juno-lan-agent
docker tag juno-conductor $DOCKER_USER/juno-conductor

docker push $DOCKER_USER/juno-lan-agent
docker push $DOCKER_USER/juno-conductor
