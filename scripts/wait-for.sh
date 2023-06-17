#!/bin/bash

set -o errexit
set -o pipefail

NAMESPACE=kai-system

if [ -n "$1" ]
  then
    NAMESPACE=$1
fi

kubectl wait --for=condition=Ready pods --all --timeout=240s -n $NAMESPACE
kubectl get pods -n $NAMESPACE
kubectl describe pods -n $NAMESPACE