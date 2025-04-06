#!/usr/bin/env bash

set -ex

MOONLANDER_VERSION=0.1.1
MOONLANDER_NAMESPACE=urmanac

helm upgrade --install karpenter-crd oci://zot.urmanac.com/moon/lander-crd \
  --version ${MOONLANDER_VERSION} --namespace "${MOONLANDER_NAMESPACE}" --create-namespace

kubectl apply -f test/yaml/test-lander.yaml
