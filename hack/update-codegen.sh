#!/usr/bin/env bash

# Copyright 2024 Observatory Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.

set -o errexit
set -o nounset
set -o pipefail

# Ensure we're in the project root
REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "${REPO_ROOT}"

echo "Updating CRD manifests..."
controller-gen crd \
  paths=./api/... \
  output:crd:dir=./config/crd

echo "Updating RBAC manifests..."
controller-gen rbac:roleName=manager-role \
  paths=./controllers/... \
  output:rbac:dir=./config/rbac

echo "Updating webhook manifests..."
controller-gen webhook \
  paths=./api/... \
  output:webhook:dir=./config/webhook

echo "Updating deepcopy code..."
controller-gen object:headerFile=./hack/boilerplate.go.txt \
  paths=./api/...

echo "✅ Code generation complete!"
