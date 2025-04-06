#!/usr/bin/env bash

set -ex

kubectl apply -k deploy/test/
kubectl wait --for=condition=available --timeout=60s deployment/moonlander -n urmanac
kubectl logs deployment/moonlander -n urmanac
