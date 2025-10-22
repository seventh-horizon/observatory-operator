# Contributing to Observatory Operator

Thank you for your interest in contributing to the Observatory Operator! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Review Process](#code-review-process)
- [Style Guidelines](#style-guidelines)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. Please be respectful and professional in all interactions.

## Getting Started

1. **Fork the Repository**: Click the "Fork" button on GitHub to create your own copy
2. **Clone Your Fork**:
   ```sh
   git clone https://github.com/YOUR-USERNAME/observatory-operator.git
   cd observatory-operator
   ```
3. **Add Upstream Remote**:
   ```sh
   git remote add upstream https://github.com/seventh-horizon/observatory-operator.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker
- kubectl
- Access to a Kubernetes cluster (kind, minikube, or cloud provider)
- Make

### Install Dependencies

```sh
go mod download
```

### Install Development Tools

The Makefile will automatically download required tools (controller-gen, kustomize, etc.) when needed:

```sh
make help
```

## Making Changes

### Create a Feature Branch

```sh
git checkout -b feature/my-new-feature
```

### Development Workflow

1. **Make your changes** to the codebase
2. **Update tests** if you're changing functionality
3. **Update documentation** if you're changing user-facing behavior
4. **Run tests locally**:
   ```sh
   make test
   ```
5. **Lint your code**:
   ```sh
   make fmt
   make vet
   ```
6. **Generate manifests** if you modified APIs:
   ```sh
   make generate
   make manifests
   ```

### Building and Testing Locally

**Build the manager binary:**

```sh
make build
```

**Run the controller locally** (requires kubeconfig):

```sh
make run
```

**Build and load Docker image** (for kind clusters):

```sh
make docker-build IMG=observatory-operator:dev
kind load docker-image observatory-operator:dev
```

## Testing

### Unit Tests

Run all unit tests:

```sh
make test
```

### Integration Tests

Integration tests run against a real Kubernetes cluster. Ensure you have access to a test cluster:

```sh
# Install CRDs
make install

# Run tests
make test
```

### End-to-End Tests

1. Deploy the operator to a test cluster:
   ```sh
   make deploy IMG=<your-test-image>
   ```

2. Apply sample resources:
   ```sh
   kubectl apply -f config/samples/
   ```

3. Verify functionality:
   ```sh
   kubectl get observatories
   kubectl describe observatory observatory-sample
   ```

4. Clean up:
   ```sh
   make undeploy
   make uninstall
   ```

## Submitting Changes

### Commit Messages

Write clear, descriptive commit messages following these guidelines:

- Use the imperative mood ("Add feature" not "Added feature")
- First line should be 50 characters or less
- Reference issues and pull requests where appropriate
- Include detailed explanation in the body if needed

Example:
```
Add retry logic to task execution

Implements exponential backoff for failed tasks. Tasks will retry
up to 3 times before being marked as failed.

Fixes #123
```

### Pull Request Process

1. **Update your branch** with the latest upstream changes:
   ```sh
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your changes** to your fork:
   ```sh
   git push origin feature/my-new-feature
   ```

3. **Create a Pull Request** on GitHub:
   - Provide a clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - List any breaking changes

4. **Address review feedback**:
   - Make requested changes
   - Push updates to the same branch
   - Respond to comments

5. **Ensure CI passes**:
   - All tests must pass
   - Code must be properly formatted
   - No linting errors

## Code Review Process

- All submissions require review from at least one maintainer
- Reviews focus on:
  - Correctness and functionality
  - Code quality and style
  - Test coverage
  - Documentation completeness
  - Backward compatibility
- Be responsive to feedback and questions
- Reviews may take a few days; please be patient

## Style Guidelines

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (run `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Run `go vet` to catch common mistakes (run `make vet`)
- Keep functions small and focused
- Write descriptive variable and function names

### API Design

When modifying APIs in `api/v1alpha1/`:

- Follow Kubernetes API conventions
- Use appropriate kubebuilder markers
- Add validation where appropriate
- Document all fields with comments
- Consider backward compatibility
- Update generated code with `make generate manifests`

### Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported functions and types
- Include examples for new features
- Update API reference documentation
- Keep CONTRIBUTING.md current

### Testing

- Write unit tests for new functionality
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error cases and edge conditions

## Project Structure

```
observatory-operator/
├── api/v1alpha1/          # API type definitions
├── cmd/                   # Main application entry point
├── config/                # Kubernetes manifests
│   ├── crd/              # CRD definitions
│   ├── default/          # Default kustomization
│   ├── manager/          # Manager deployment
│   ├── rbac/             # RBAC configuration
│   └── samples/          # Sample resources
├── internal/controller/   # Controller implementation
├── hack/                 # Build scripts and tools
├── Dockerfile            # Container image definition
├── Makefile              # Build automation
└── README.md             # Project documentation
```

## Adding New Features

### Adding a New API Field

1. Modify the type definition in `api/v1alpha1/*_types.go`
2. Add kubebuilder validation markers
3. Run `make generate manifests`
4. Update sample resources in `config/samples/`
5. Update documentation
6. Add tests

### Adding a New Controller

1. Create controller file in `internal/controller/`
2. Implement Reconcile method
3. Add RBAC markers
4. Register controller in `cmd/main.go`
5. Run `make manifests`
6. Add tests
7. Update documentation

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open a GitHub Issue
- **Features**: Open a GitHub Issue with the feature request template
- **Security**: See SECURITY.md for reporting security issues

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for their contributions
- Project documentation

Thank you for contributing to Observatory Operator!
