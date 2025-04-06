#!/usr/bin/env bash

set -ex

kubectl delete -k deploy/test/ || true
kubectl delete -f test/yaml/ || true
helm uninstall -n urmanac moonlander-crd
