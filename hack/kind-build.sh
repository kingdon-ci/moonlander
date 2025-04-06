#!/usr/bin/env bash

DEPLOYMENT_PATH="deploy/test/deployment.yaml"

set -x

export KO_DOCKER_REPO=kind.local
ko build ./cmd/moonlander --local --bare > image.txt
IMAGE=$(cat image.txt)
echo "Built image: $IMAGE"
kind load docker-image $IMAGE
echo "Patching $DEPLOYMENT_PATH with image $IMAGE"
yq e -i ".spec.template.spec.containers[0].image = \"${IMAGE}\"" $DEPLOYMENT_PATH
