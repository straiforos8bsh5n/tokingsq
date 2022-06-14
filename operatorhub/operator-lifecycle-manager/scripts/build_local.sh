#!/usr/bin/env bash

# Note: run from root
# This is used to start and build services for running e2e tests

set -e

ps x | grep -q [m]inikube || minikube start --kubernetes-version="v1.11.1" --extra-config=apiserver.v=4 || { echo 'Cannot start minikube.'; exit 1; }
eval $(minikube docker-env) || { echo 'Cannot switch to minikube docker'; exit 1; }
kubectl config use-context minikube
docker build -f e2e.Dockerfile .
docker tag $(docker images --filter 'label=stage=olm' --format '{{.CreatedAt}}\t{{.ID}}' | sort -nr | head -n 1 | cut -f2) quay.io/coreos/olm:local
docker tag $(docker images --filter 'label=stage=e2e' --format '{{.CreatedAt}}\t{{.ID}}' | sort -nr | head -n 1 | cut -f2) quay.io/coreos/olm-e2e:local
