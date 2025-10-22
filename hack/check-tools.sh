#!/usr/bin/env bash

# Copyright 2024 Observatory Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.

set -o errexit
set -o nounset
set -o pipefail

echo "Checking required tools..."

REQUIRED_TOOLS=(
    "go:Go compiler"
    "kubectl:Kubernetes CLI"
    "kustomize:Kustomize"
    "controller-gen:Controller generator"
    "docker:Docker"
)

MISSING=()

for tool_spec in "${REQUIRED_TOOLS[@]}"; do
    IFS=':' read -r tool desc <<< "$tool_spec"
    if ! command -v "$tool" &> /dev/null; then
        MISSING+=("$tool ($desc)")
        echo "❌ $tool not found"
    else
        version=$($tool version 2>/dev/null | head -n1 || echo "unknown")
        echo "✅ $tool: $version"
    fi
done

if [ ${#MISSING[@]} -ne 0 ]; then
    echo ""
    echo "Missing required tools:"
    for tool in "${MISSING[@]}"; do
        echo "  - $tool"
    done
    echo ""
    echo "Installation instructions:"
    echo "  Go: https://golang.org/doc/install"
    echo "  kubectl: https://kubernetes.io/docs/tasks/tools/"
    echo "  kustomize: https://kubectl.docs.kubernetes.io/installation/kustomize/"
    echo "  controller-gen: go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest"
    echo "  docker: https://docs.docker.com/get-docker/"
    exit 1
fi

echo ""
echo "✅ All required tools are installed!"
