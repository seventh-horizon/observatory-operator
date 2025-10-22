# Observatory Operator

A Kubernetes operator for managing DAG-based workflows via Custom Resource Definitions (CRDs). The Observatory Operator enables you to run reproducible, dependency-aware pipelines with built-in support for metrics, alerts, and OpenTelemetry (OTEL) integration.

## Description

The Observatory Operator extends Kubernetes with the ability to orchestrate complex, multi-step workflows as directed acyclic graphs (DAGs). Each workflow is represented as an `Observatory` custom resource, which defines a collection of tasks with their dependencies.

### Key Features

- **DAG-based Workflows**: Define workflows as directed acyclic graphs with explicit task dependencies
- **Declarative Configuration**: Workflows are defined using Kubernetes-native YAML manifests
- **Dependency Management**: Tasks automatically execute in the correct order based on their dependencies
- **Concurrent Execution**: Multiple independent tasks can run in parallel
- **Scheduling Support**: Optional cron-based scheduling for recurring workflows
- **Status Tracking**: Real-time status updates for workflows and individual tasks
- **Kubernetes Native**: Built with kubebuilder and controller-runtime for seamless integration

## Getting Started

### Prerequisites

- Go 1.21+
- Kubernetes cluster (v1.28+)
- kubectl configured to communicate with your cluster
- Docker (for building container images)

### Installation

1. **Install the CRDs into the cluster:**

```sh
make install
```

2. **Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<your-registry>/observatory-operator:tag
```

3. **Deploy the controller to the cluster:**

```sh
make deploy IMG=<your-registry>/observatory-operator:tag
```

### Quick Start Example

Create a simple workflow with dependent tasks:

```yaml
apiVersion: workflow.seventh-horizon.io/v1alpha1
kind: Observatory
metadata:
  name: my-workflow
spec:
  tasks:
    - name: setup
      image: busybox:1.36
      command: ["sh", "-c"]
      args: ["echo 'Setup complete'"]
    
    - name: process
      image: busybox:1.36
      command: ["sh", "-c"]
      args: ["echo 'Processing data'"]
      dependencies:
        - setup
    
    - name: finalize
      image: busybox:1.36
      command: ["sh", "-c"]
      args: ["echo 'Workflow complete'"]
      dependencies:
        - process
```

Apply the workflow:

```sh
kubectl apply -f my-workflow.yaml
```

Check the workflow status:

```sh
kubectl get observatory my-workflow
kubectl describe observatory my-workflow
```

## Architecture

The Observatory Operator consists of:

1. **Custom Resource Definition (CRD)**: Defines the `Observatory` API
2. **Controller**: Watches `Observatory` resources and reconciles them
3. **Reconciliation Loop**: Manages the lifecycle of workflow tasks

### Workflow Execution

1. User creates an `Observatory` custom resource
2. Controller detects the new resource
3. Controller analyzes task dependencies to build execution graph
4. Tasks with no dependencies or satisfied dependencies are scheduled
5. As tasks complete, dependent tasks become eligible for execution
6. Controller updates status throughout the workflow lifecycle

## Development

### Running Locally

**Run against a Kubernetes cluster:**

```sh
make run
```

**Create sample Observatory:**

```sh
kubectl apply -f config/samples/workflow_v1alpha1_observatory.yaml
```

### Testing

**Run unit tests:**

```sh
make test
```

### Building

**Build the manager binary:**

```sh
make build
```

**Build the Docker image:**

```sh
make docker-build IMG=<your-registry>/observatory-operator:tag
```

### Code Generation

After modifying the API definitions in `api/v1alpha1/`, regenerate code and manifests:

```sh
make generate
make manifests
```

## API Reference

### Observatory Spec

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `tasks` | `[]TaskSpec` | List of tasks in the workflow | Yes |
| `schedule` | `string` | Cron schedule for recurring execution | No |
| `suspend` | `bool` | Suspend workflow execution | No |
| `maxConcurrent` | `int32` | Maximum concurrent task executions | No |

### TaskSpec

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | `string` | Unique task identifier | Yes |
| `image` | `string` | Container image to run | Yes |
| `command` | `[]string` | Container command | No |
| `args` | `[]string` | Container arguments | No |
| `dependencies` | `[]string` | Tasks that must complete first | No |
| `env` | `[]EnvVar` | Environment variables | No |

### Observatory Status

| Field | Type | Description |
|-------|------|-------------|
| `phase` | `string` | Current workflow phase (Pending, Running, Succeeded, Failed) |
| `tasks` | `[]TaskStatus` | Status of individual tasks |
| `startTime` | `*metav1.Time` | When workflow started |
| `completionTime` | `*metav1.Time` | When workflow completed |
| `conditions` | `[]metav1.Condition` | Detailed status conditions |

## Configuration

### Environment Variables

The operator manager supports the following command-line flags:

- `--metrics-bind-address`: Address for metrics endpoint (default: `:8080`)
- `--health-probe-bind-address`: Address for health probe endpoint (default: `:8081`)
- `--leader-elect`: Enable leader election for high availability (default: `false`)
- `--metrics-secure`: Serve metrics securely (default: `false`)
- `--enable-http2`: Enable HTTP/2 for metrics and webhooks (default: `false`)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Monitoring and Observability

The operator exposes Prometheus metrics at `/metrics` endpoint. Metrics include:

- Controller reconciliation metrics
- Workflow execution counts
- Task success/failure rates
- Reconciliation latency

## Troubleshooting

### Common Issues

**Workflow stuck in Pending:**
- Check if tasks have circular dependencies
- Verify all container images are accessible
- Check controller logs: `kubectl logs -n observatory-operator-system deployment/observatory-operator-controller-manager`

**Tasks not executing:**
- Verify RBAC permissions are correctly configured
- Check pod events: `kubectl get events -n <namespace>`
- Review task specifications for errors

## Uninstall

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy Controller

To remove the controller from the cluster:

```sh
make undeploy
```

## License

Copyright 2024 The Observatory Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Support

For issues and questions:
- Open an issue on GitHub
- Check existing documentation
- Review sample configurations in `config/samples/`

## Roadmap

- [ ] Enhanced scheduling capabilities
- [ ] Webhook validation for Observatory resources
- [ ] Advanced retry and failure handling strategies
- [ ] Integration with external secret management
- [ ] Support for workflow templates
- [ ] Workflow versioning and rollback
- [ ] Enhanced observability with OpenTelemetry
- [ ] Multi-cluster workflow execution
