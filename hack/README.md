# Hack Scripts

This directory contains utility scripts for development, testing, and code generation.

## Scripts Overview

| Script               | Purpose                                          | Usage                           |
| -------------------- | ------------------------------------------------ | ------------------------------- |
| `update-codegen.sh`  | Generate CRDs, RBAC, webhooks, and deepcopy code | `./hack/update-codegen.sh`      |
| `verify-codegen.sh`  | Verify generated code is up to date (CI)         | `./hack/verify-codegen.sh`      |
| `check-tools.sh`     | Verify all required tools are installed          | `./hack/check-tools.sh`         |
| `setup-local.sh`     | Setup local kind cluster for development         | `./hack/setup-local.sh`         |
| `run-tests.sh`       | Run all tests with options                       | `./hack/run-tests.sh [options]` |
| `boilerplate.go.txt` | License header template for generated files      | (used by controller-gen)        |

## Making Scripts Executable

After cloning, make the scripts executable:

```bash
chmod +x hack/*.sh
```

## Script Details

### update-codegen.sh

Regenerates all Kubernetes manifests and Go code from API definitions.

**Generates:**

- CRD manifests in `config/crd/`
- RBAC manifests in `config/rbac/`
- Webhook manifests in `config/webhook/`
- DeepCopy methods in `api/`

**When to run:**

- After modifying API types in `api/v1alpha1/`
- After changing controller RBAC annotations
- Before committing changes

**Example:**

```bash
./hack/update-codegen.sh
```

### verify-codegen.sh

Checks if generated code matches the current API definitions. Used in CI to ensure developers ran `update-codegen.sh`.

**Example:**

```bash
./hack/verify-codegen.sh
# Exit code 0 = up to date
# Exit code 1 = needs regeneration
```

**In CI (.github/workflows/ci.yml):**

```yaml
- name: Verify codegen
  run: ./hack/verify-codegen.sh
```

### check-tools.sh

Validates that all required development tools are installed and shows their versions.

**Checks for:**

- Go compiler
- kubectl
- kustomize
- controller-gen
- Docker

**Example:**

```bash
./hack/check-tools.sh

# Output:
✅ go: go version go1.21.0 linux/amd64
✅ kubectl: Client Version: v1.28.0
✅ kustomize: v5.2.1
✅ controller-gen: Version: v0.13.0
✅ docker: Docker version 24.0.6
```

### setup-local.sh

Creates a local Kubernetes cluster using kind with:

- Local Docker registry (localhost:5001)
- cert-manager installed
- Proper network configuration

**Environment variables:**

- `KIND_CLUSTER_NAME`: Cluster name (default: `observatory-dev`)
- `REGISTRY_NAME`: Registry container name (default: `kind-registry`)
- `REGISTRY_PORT`: Local registry port (default: `5001`)

**Example:**

```bash
# Create default cluster
./hack/setup-local.sh

# Create custom cluster
KIND_CLUSTER_NAME=my-test ./hack/setup-local.sh

# Build and deploy
make docker-build docker-push IMG=localhost:5001/observatory-operator:dev
make deploy IMG=localhost:5001/observatory-operator:dev

# Cleanup
kind delete cluster --name observatory-dev
```

### run-tests.sh

Runs all tests with various options.

**Options:**

- `-c, --coverage`: Generate coverage report
- `-v, --verbose`: Verbose test output
- `-r, --run PATTERN`: Run specific tests matching pattern

**Examples:**

```bash
# Run all tests
./hack/run-tests.sh

# Run with coverage
./hack/run-tests.sh --coverage

# Run specific test
./hack/run-tests.sh --run TestReconcile

# Verbose output with coverage
./hack/run-tests.sh -v -c
```

**Coverage report:**
After running with `--coverage`, view the HTML report:

```bash
go tool cover -html=coverage.out
```

## Development Workflow

### Initial Setup

```bash
# 1. Check tools
./hack/check-tools.sh

# 2. Setup local cluster
./hack/setup-local.sh

# 3. Run tests
./hack/run-tests.sh
```

### Making Changes

```bash
# 1. Modify API types in api/v1alpha1/
vim api/v1alpha1/observatoryrun_types.go

# 2. Regenerate code
./hack/update-codegen.sh

# 3. Run tests
./hack/run-tests.sh -c

# 4. Verify everything
./hack/verify-codegen.sh
```

### Before Committing

```bash
# Run full verification
./hack/verify-codegen.sh
./hack/run-tests.sh --coverage
go fmt ./...
go vet ./...
```

## CI Integration

These scripts are used in GitHub Actions workflows:

**`.github/workflows/ci.yml`:**

```yaml
name: CI

on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.21"

      - name: Check tools
        run: ./hack/check-tools.sh

      - name: Verify codegen
        run: ./hack/verify-codegen.sh

      - name: Run tests
        run: ./hack/run-tests.sh --coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

## Troubleshooting

### "controller-gen: command not found"

Install controller-gen:

```bash
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

### "kind: command not found"

Install kind:

```bash
# Linux/macOS
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Or use package manager
# macOS: brew install kind
# Linux: See https://kind.sigs.k8s.io/docs/user/quick-start/
```

### "setup-envtest: command not found"

Install setup-envtest:

```bash
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
```

### Permission Denied

Make scripts executable:

```bash
chmod +x hack/*.sh
```

## Adding New Scripts

When adding new scripts:

1. Use the boilerplate header:

```bash
#!/usr/bin/env bash

# Copyright 2024 Observatory Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.

set -o errexit
set -o nounset
set -o pipefail
```

2. Add to this README
3. Make executable: `chmod +x hack/new-script.sh`
4. Test in both development and CI environments

## References

- [Kubebuilder Book - Testing](https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html)
- [Controller Runtime - Testing](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)
- [Kind - Local Clusters](https://kind.sigs.k8s.io/)
- [Kustomize](https://kustomize.io/)
