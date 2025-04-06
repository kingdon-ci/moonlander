#!/usr/bin/env bash

set -ex

KARPENTER_VERSION=1.3.3
KARPENTER_NAMESPACE=kube-system

helm upgrade --install karpenter-crd oci://zot.urmanac.com/moon/lander-crd \
  --version ${KARPENTER_VERSION} --namespace "${KARPENTER_NAMESPACE}" --create-namespace

kubectl apply -f test/yaml/test-lander.yaml
