# Observatory Operator (Starter, with Webhook + Controller)
![Telemetry Contract](https://github.com/seventh-horizon/observatory-operator/actions/workflows/telemetry.yml/badge.svg?branch=main)

A minimal-but-functional Kubernetes operator that runs simple DAG-based workflows by creating Kubernetes Jobs per task. Includes:

- ✅ Validating admission webhook (cycle detection, dependency checks, sane warnings)
- ✅ Idempotent controller that creates Jobs and tracks status
- ✅ Samples (sequential, DAG)
- ✅ RBAC, manager deployment, cert-manager integration
- ✅ Makefile and kustomize wiring

## Quick Start

```bash
make controller-gen kustomize
make manifests

# create dev cluster (optional)
make kind-create

# install CRDs, RBAC, manager+webhook (requires cert-manager installed in your cluster)
kubectl create ns observatory-system || true
make install
# install cert-manager separately if you don't have it
# kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# deploy controller (set IMG to your pushed image if using a real cluster)
make deploy IMG=observatory-operator:latest
```

Run an example:

```bash
make run-example-simple
kubectl get observatoryruns -w
kubectl get jobs
```

## Validation Errors Examples

- Circular dependency is rejected
- Missing dependencies are rejected
- Invalid task names rejected

See `config/samples/invalid-circular.yaml` for a quick test.

## Developer Notes
See: docs/DEVELOPER_NOTES.md


## Developer Notes
See: docs/DEVELOPER_NOTES.md


## Developer Notes
See: docs/DEVELOPER_NOTES.md

