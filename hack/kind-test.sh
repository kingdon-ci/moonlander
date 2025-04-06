#!/usr/bin/env bash

set -x

go test -v ./test/integration -count=1
kubectl logs deployment/moonlander -n urmanac
