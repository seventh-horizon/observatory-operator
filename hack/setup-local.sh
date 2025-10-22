#!/usr/bin/env bash

# Copyright 2024 Observatory Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "${REPO_ROOT}"

KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-observatory-dev}"
REGISTRY_NAME="${REGISTRY_NAME:-kind-registry}"
REGISTRY_PORT="${REGISTRY_PORT:-5001}"

echo "🚀 Setting up local development environment..."

# Check if kind cluster exists
if ! kind get clusters | grep -q "^${KIND_CLUSTER_NAME}$"; then
    echo "Creating kind cluster: ${KIND_CLUSTER_NAME}"
    cat <<EOF | kind create cluster --name "${KIND_CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REGISTRY_PORT}"]
    endpoint = ["http://${REGISTRY_NAME}:5000"]
EOF
else
    echo "✅ Kind cluster already exists: ${KIND_CLUSTER_NAME}"
fi

# Start local registry if not running
if ! docker ps | grep -q "${REGISTRY_NAME}"; then
    echo "Starting local registry: ${REGISTRY_NAME}"
    docker run -d --restart=always \
        -p "127.0.0.1:${REGISTRY_PORT}:5000" \
        --name "${REGISTRY_NAME}" \
        registry:2
    
    # Connect registry to kind network
    docker network connect "kind" "${REGISTRY_NAME}" 2>/dev/null || true
else
    echo "✅ Local registry already running: ${REGISTRY_NAME}"
fi

# Install cert-manager
echo "Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager
kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook

echo ""
echo "✅ Local development environment ready!"
echo ""
echo "Next steps:"
echo "  1. Build and load image: make docker-build docker-push IMG=localhost:${REGISTRY_PORT}/observatory-operator:dev"
echo "  2. Deploy: make deploy IMG=localhost:${REGISTRY_PORT}/observatory-operator:dev"
echo "  3. Test: kubectl apply -f config/samples/simple-workflow.yaml"
echo ""
echo "To delete the cluster: kind delete cluster --name ${KIND_CLUSTER_NAME}"
