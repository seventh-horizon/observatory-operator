# OpenTelemetry Collector Configuration

This directory contains the OpenTelemetry Collector configuration for the Observatory Operator.

## Overview

The OTEL collector provides a unified pipeline for:

- **Metrics**: Prometheus metrics from the controller
- **Traces**: Distributed tracing for workflow execution
- **Logs**: Structured logs and Kubernetes events

## Quick Start

### 1. Deploy the Collector

```bash
kubectl apply -k config/otel/
```

### 2. Configure Your Backend

Edit `collector-config.yaml` to configure your observability backend:

**For Prometheus + Grafana:**

```yaml
exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
```

**For Cloud Providers:**

Uncomment the appropriate exporter in `collector-config.yaml`:

- Google Cloud: `googlecloud`
- AWS CloudWatch: `awscloudwatch`
- Azure Monitor: `azuremonitor`
- Datadog: `datadog`

### 3. Enable OTEL in Controller

Set environment variable in the controller deployment:

```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "http://otel-collector:4317"
```

## Configuration Files

- **collector-config.yaml**: Main collector configuration
- **deployment.yaml**: Kubernetes deployment manifest
- **kustomization.yaml**: Kustomize configuration

## Architecture

```
┌─────────────────────────┐
│ Observatory Controller  │
│  - Prometheus metrics   │───┐
│  - OTLP traces         │   │
│  - Structured logs     │   │
└─────────────────────────┘   │
                              │
           ┌──────────────────┘
           │
           ▼
┌─────────────────────────┐
│   OTEL Collector        │
│  ┌─────────────────┐    │
│  │   Receivers     │    │
│  │ - Prometheus    │    │
│  │ - OTLP          │    │
│  │ - K8s Events    │    │
│  └────────┬────────┘    │
│           │             │
│  ┌────────▼────────┐    │
│  │   Processors    │    │
│  │ - Batch         │    │
│  │ - Filter        │    │
│  │ - Transform     │    │
│  └────────┬────────┘    │
│           │             │
│  ┌────────▼────────┐    │
│  │   Exporters     │    │
│  │ - Prometheus    │────┼───▶ Prometheus
│  │ - Jaeger        │────┼───▶ Jaeger
│  │ - Loki          │────┼───▶ Loki
│  │ - Cloud         │────┼───▶ Cloud Backend
│  └─────────────────┘    │
└─────────────────────────┘
```

## Metrics Collected

The collector scrapes these metrics from the Observatory controller:

| Metric                                   | Type      | Description                   |
| ---------------------------------------- | --------- | ----------------------------- |
| `observatory_reconcile_total`            | Counter   | Total reconciliation attempts |
| `observatory_reconcile_duration_seconds` | Histogram | Reconciliation latency        |
| `observatory_reconcile_errors_total`     | Counter   | Reconciliation errors         |
| `observatory_workflow_duration_seconds`  | Histogram | Workflow execution time       |
| `observatory_workflow_retries_total`     | Counter   | Workflow retries              |
| `observatory_job_*`                      | Counter   | Job lifecycle metrics         |
| `observatory_active_workflows`           | Gauge     | Active workflows              |

## Sampling and Filtering

**Trace Sampling:**

- 10% of traces are sampled by default
- Adjust `probabilistic_sampler.sampling_percentage`

**Metric Filtering:**

- Test and debug metrics are filtered out
- Modify `filter/metrics` processor to customize

## Resource Limits

Default resource limits:

- CPU: 100m request, 500m limit
- Memory: 128Mi request, 512Mi limit

Adjust based on your telemetry volume.

## Health Checks

The collector exposes health endpoints:

- `/` on port 13133 - Liveness/readiness
- `/metrics` on port 8888 - Collector's own metrics
- `/debug/zpages` on port 55679 - Live debugging

## Troubleshooting

### No Metrics Appearing

1. Check collector logs:

```bash
kubectl logs -n observatory-system deployment/otel-collector
```

2. Verify controller is reachable:

```bash
kubectl exec -n observatory-system deployment/otel-collector -- \
  wget -O- http://observatory-controller-metrics:8080/metrics
```

### High Memory Usage

1. Reduce batch size in `batch` processor
2. Increase `check_interval` in `memory_limiter`
3. Enable more aggressive filtering

### Collector Not Starting

Check ConfigMap is mounted correctly:

```bash
kubectl describe pod -n observatory-system -l app=otel-collector
```

## Advanced Configuration

### Custom Exporters

Add custom exporters to `collector-config.yaml`:

```yaml
exporters:
  custom_backend:
    endpoint: https://your-backend.com
    headers:
      Authorization: Bearer ${API_TOKEN}

service:
  pipelines:
    metrics:
      exporters: [prometheus, custom_backend]
```

### Multiple Pipelines

Split metrics by label:

```yaml
processors:
  filter/critical:
    metrics:
      include:
        match_type: regexp
        metric_names:
          - "observatory_(error|timeout).*"

service:
  pipelines:
    metrics/critical:
      receivers: [prometheus]
      processors: [filter/critical]
      exporters: [pagerduty]
```

## References

- [OTEL Collector Documentation](https://opentelemetry.io/docs/collector/)
- [Prometheus Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver)
- [Available Exporters](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter)
