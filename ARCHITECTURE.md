# Observatory Operator Architecture

## Overview

The Observatory Operator is a Kubernetes operator that manages DAG-based workflows through custom resources. It follows the Kubernetes operator pattern using the controller-runtime framework.

## Components

### 1. Custom Resource Definition (CRD)

**File:** `config/crd/bases/workflow.seventh-horizon.io_observatories.yaml`

The Observatory CRD defines the API for workflow resources:

```yaml
apiVersion: workflow.seventh-horizon.io/v1alpha1
kind: Observatory
```

**Key Features:**
- Namespaced scope
- Status subresource for tracking workflow state
- Additional printer columns for `kubectl` output (Phase, Tasks, Age)
- Comprehensive validation rules

### 2. API Types

**Location:** `api/v1alpha1/`

#### ObservatorySpec
Defines the desired state of a workflow:
- `tasks`: List of TaskSpec defining the workflow DAG
- `schedule`: Optional cron schedule for recurring execution
- `suspend`: Flag to pause workflow execution
- `maxConcurrent`: Limit on concurrent task executions

#### TaskSpec
Defines a single task in the workflow:
- `name`: Unique identifier
- `image`: Container image to execute
- `command` / `args`: Container command and arguments
- `dependencies`: List of task names that must complete first
- `env`: Environment variables

#### ObservatoryStatus
Tracks the observed state:
- `phase`: Overall workflow state (Pending, Running, Succeeded, Failed)
- `tasks`: Status of individual tasks
- `startTime` / `completionTime`: Timing information
- `conditions`: Detailed status conditions

### 3. Controller

**File:** `internal/controller/observatory_controller.go`

The controller implements the reconciliation logic:

```go
func (r *ObservatoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
```

**Responsibilities:**
1. Watch Observatory resources
2. Parse DAG structure from task dependencies
3. Create pods for tasks ready to execute
4. Monitor task execution
5. Update workflow status
6. Handle failures and retries

**RBAC Permissions:**
- Full access to Observatory resources
- Pod management (create, delete, get, list, watch, update, patch)
- Status updates for Observatories

### 4. Manager (Main)

**File:** `cmd/main.go`

The main entry point that:
1. Initializes the Kubernetes client
2. Registers the API scheme
3. Sets up the controller manager
4. Configures metrics and health endpoints
5. Enables leader election for high availability

## Deployment Architecture

```
┌─────────────────────────────────────────────────┐
│           Kubernetes Cluster                     │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │  observatory-operator-system namespace     │ │
│  │                                            │ │
│  │  ┌──────────────────────────────────────┐ │ │
│  │  │  Observatory Operator Manager        │ │ │
│  │  │  - Watches Observatory CRs           │ │ │
│  │  │  - Reconciles workflow state         │ │ │
│  │  │  - Creates task pods                 │ │ │
│  │  │  - Updates status                    │ │ │
│  │  └──────────────────────────────────────┘ │ │
│  │                                            │ │
│  │  ┌──────────────────────────────────────┐ │ │
│  │  │  kube-rbac-proxy (optional)          │ │ │
│  │  │  - Secures metrics endpoint          │ │ │
│  │  └──────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────┘ │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │  User namespaces                           │ │
│  │                                            │ │
│  │  ┌──────────────────────────────────────┐ │ │
│  │  │  Observatory CR: my-workflow         │ │ │
│  │  │  - Defines tasks and dependencies    │ │ │
│  │  │  - Status tracked by operator        │ │ │
│  │  └──────────────────────────────────────┘ │ │
│  │                                            │ │
│  │  ┌──────────────────────────────────────┐ │ │
│  │  │  Task Pods (created by operator)     │ │ │
│  │  │  - pod/my-workflow-task-1            │ │ │
│  │  │  - pod/my-workflow-task-2            │ │ │
│  │  │  - pod/my-workflow-task-3            │ │ │
│  │  └──────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

## Workflow Execution Flow

1. **User Creates Observatory CR**
   ```
   kubectl apply -f my-workflow.yaml
   ```

2. **Controller Detects New Resource**
   - Reconcile loop triggered
   - Observatory object fetched from API server

3. **DAG Analysis**
   - Parse task dependencies
   - Build execution graph
   - Identify tasks ready to run (no dependencies or all dependencies satisfied)

4. **Task Scheduling**
   - Create Pod for each ready task
   - Set appropriate labels and owner references
   - Configure pod spec from TaskSpec

5. **Status Monitoring**
   - Watch task pods for state changes
   - Update TaskStatus for each task
   - Update overall Observatory status

6. **Dependency Resolution**
   - As tasks complete, check dependent tasks
   - Schedule newly eligible tasks
   - Handle failures and retries

7. **Completion**
   - All tasks succeeded → Observatory phase: Succeeded
   - Any task failed → Observatory phase: Failed
   - Update completion time

## Configuration Structure

```
config/
├── crd/               # Custom Resource Definitions
│   └── bases/         # Generated CRD manifests
├── default/           # Default kustomization
│   └── kustomization.yaml
├── manager/           # Operator deployment
│   ├── manager.yaml   # Deployment manifest
│   └── controller_manager_config.yaml
├── rbac/              # Role-Based Access Control
│   ├── role.yaml      # ClusterRole for operator
│   ├── role_binding.yaml
│   └── service_account.yaml
└── samples/           # Example Observatory resources
    └── workflow_v1alpha1_observatory.yaml
```

## Build and Development Tools

### Makefile Targets

- `make build`: Build the manager binary
- `make run`: Run the operator locally
- `make docker-build`: Build container image
- `make deploy`: Deploy to Kubernetes cluster
- `make install`: Install CRDs
- `make uninstall`: Remove CRDs
- `make test`: Run tests
- `make fmt`: Format code
- `make vet`: Run go vet

### Code Generation

The operator uses controller-gen for code generation:
- **DeepCopy methods**: Generated in `zz_generated.deepcopy.go`
- **CRD manifests**: Generated from Go type annotations
- **RBAC manifests**: Generated from controller annotations

## Security Considerations

1. **RBAC**
   - Principle of least privilege
   - Separate service account for operator
   - ClusterRole for cluster-wide resources
   - Role for namespace-scoped resources

2. **Pod Security**
   - Non-root containers
   - Read-only root filesystem
   - No privilege escalation
   - Dropped capabilities

3. **Metrics Security**
   - kube-rbac-proxy protects metrics endpoint
   - RBAC-based access control
   - TLS support

## Extensibility Points

### Custom Validation
Add webhook validation in `api/v1alpha1/observatory_webhook.go`

### Custom Controllers
Add additional controllers for related resources

### Status Conditions
Extend status with custom condition types

### Metrics
Add custom Prometheus metrics in the controller

## Future Enhancements

1. **Retry Logic**: Automatic task retry on failure
2. **Timeouts**: Task-level timeout configuration
3. **Parallelism**: Fine-grained control over concurrent execution
4. **Artifacts**: Task output artifact management
5. **Notifications**: Integration with notification systems
6. **Workflows as Dependencies**: Reference other Observatories
7. **Resource Limits**: CPU/memory constraints per task
8. **Conditional Execution**: Skip tasks based on conditions
9. **Dynamic DAGs**: Generate task lists dynamically
10. **Audit Logging**: Track all workflow executions

## References

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Controller Runtime](https://github.com/kubernetes-sigs/controller-runtime)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
