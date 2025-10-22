# Changelog

All notable changes to the Observatory Operator project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Comprehensive Prometheus metrics for observability
  - `observatory_reconcile_total` - Total reconciliation attempts
  - `observatory_reconcile_duration_seconds` - Reconciliation duration histogram
  - `observatory_reconcile_errors_total` - Total reconciliation errors
  - `observatory_workflow_duration_seconds` - Workflow execution time
  - `observatory_workflow_retries_total` - Workflow retry counter
  - `observatory_job_create_errors_total` - Job creation failures
  - `observatory_job_completed_total` - Completed jobs counter
  - `observatory_job_timeout_total` - Job timeout counter
  - `observatory_webhook_errors_total` - Webhook validation errors
  - `observatory_active_workflows` - Active workflow gauge
  - `observatory_workflow_step_duration_seconds` - Per-step duration
  - `observatory_dependency_resolution_duration_seconds` - DAG resolution time

- Retry logic with exponential backoff
  - Configurable retry count per workflow step
  - Backoff starts at 30s, max 5 minutes
  - Per-step retry tracking

- PrometheusRule with SLO-based alerts
  - High reconciliation latency alert (>60s p99.9)
  - High error rate alert (>1%)
  - Job creation failure alert
  - High retry rate warning
  - Workflow slowdown detection
  - Controller health monitoring
  - Job timeout rate tracking
  - Webhook failure detection

- ServiceMonitor for Prometheus metrics scraping
- Enhanced error handling with structured logging
- Resource limit enforcement for Jobs
- Test suite with envtest integration
- OTEL collector configuration examples

### Changed

- Improved controller reconciliation loop with better status updates
- Enhanced webhook validation with circular dependency detection
- Better resource cleanup on workflow failure

### Fixed

- Memory leaks in long-running workflows
- Race conditions in concurrent job creation
- Status update conflicts during high reconciliation rates

### Security

- Added security contexts to Job pods (non-root, read-only filesystem)
- Implemented proper RBAC with minimal required permissions
- Added NetworkPolicies for controller pod

## [0.1.0] - 2024-01-15

### Added

- Initial release of Observatory Operator
- Custom Resource Definition (ObservatoryRun) for workflow execution
- Basic workflow orchestration with DAG support
- Webhook validation for CRD
- Leader election for HA deployments
- Basic RBAC configuration
- Sample workflows (simple, DAG, with-retries)
- Python telemetry dashboard
- Timeline visualization

### Documentation

- README with quick start guide
- DEVELOPMENT.md for contributors
- OPERATIONS.md for cluster operators
- API reference documentation
- Meta-observability contract specification

[Unreleased]: https://github.com/example/observatory-operator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/example/observatory-operator/releases/tag/v0.1.0
