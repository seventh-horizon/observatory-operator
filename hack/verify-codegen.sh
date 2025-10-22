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

echo "Verifying code generation is up to date..."

# Create temporary directory
TMPDIR=$(mktemp -d)
trap 'rm -rf ${TMPDIR}' EXIT

# Copy current generated files
cp -r config/crd "${TMPDIR}/"
cp -r config/rbac "${TMPDIR}/"
cp -r config/webhook "${TMPDIR}/"
find api -name "zz_generated.*.go" -exec cp {} "${TMPDIR}/" \;

# Regenerate
./hack/update-codegen.sh

# Compare
if ! diff -Naupr "${TMPDIR}/crd" config/crd/ || \
   ! diff -Naupr "${TMPDIR}/rbac" config/rbac/ || \
   ! diff -Naupr "${TMPDIR}/webhook" config/webhook/ || \
   ! find api -name "zz_generated.*.go" -exec diff -Naupr "${TMPDIR}/{}" {} \;; then
    echo "❌ Generated code is out of date!"
    echo "Run: make generate manifests"
    exit 1
fi

echo "✅ Generated code is up to date!"
