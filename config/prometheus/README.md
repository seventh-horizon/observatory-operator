# Prometheus Monitoring Configuration

This directory contains Prometheus monitoring resources for the Observatory Operator.

## Files

- **servicemonitor.yaml**: ServiceMonitor to scrape metrics from the controller
- **prometheusrule.yaml**: Alerting rules based on SLOs
- **kustomization.yaml**: Kustomize configuration

## Prerequisites

These configurations require:

- **Prometheus Operator** installed in your cluster
- **ServiceMonitor CRD** (`monitoring.coreos.com/v1`)
- **PrometheusRule CRD** (`monitoring.coreos.com/v1`)

### Installing Prometheus Operator

```bash
# Using Helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace

# Or using kubectl
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml
```

## Deployment

### Deploy with Main Operator

Add to `config/default/kustomization.yaml`:

```yaml
resources:
  - ../prometheus
```

Then deploy:

```bash
kubectl apply -k config/default/
```

### Deploy Separately

```bash
kubectl apply -k config/prometheus/
```

## Metrics Collected

The ServiceMonitor scrapes these metrics from the controller:

| Metric                                               | Type      | Labels            | Description                                   |
| ---------------------------------------------------- | --------- | ----------------- | --------------------------------------------- |
| `observatory_reconcile_total`                        | Counter   | `result`          | Total reconciliation attempts (success/error) |
| `observatory_reconcile_duration_seconds`             | Histogram | -                 | Time spent in reconciliation loop             |
| `observatory_reconcile_errors_total`                 | Counter   | `error_type`      | Total reconciliation errors by type           |
| `observatory_workflow_duration_seconds`              | Histogram | `workflow_name`   | End-to-end workflow execution time            |
| `observatory_workflow_retries_total`                 | Counter   | `step_name`       | Workflow step retry counter                   |
| `observatory_job_create_errors_total`                | Counter   | -                 | Failed Job creation attempts                  |
| `observatory_job_completed_total`                    | Counter   | `status`          | Completed jobs (success/failure)              |
| `observatory_job_timeout_total`                      | Counter   | -                 | Jobs that exceeded timeout                    |
| `observatory_webhook_errors_total`                   | Counter   | `validation_type` | Webhook validation failures                   |
| `observatory_active_workflows`                       | Gauge     | -                 | Number of currently running workflows         |
| `observatory_workflow_step_duration_seconds`         | Histogram | `step_name`       | Duration of individual workflow steps         |
| `observatory_dependency_resolution_duration_seconds` | Histogram | -                 | Time to resolve DAG dependencies              |

## Alerting Rules

The PrometheusRule defines 8 SLO-based alerts:

### Critical Alerts (severity: critical)

1. **ObservatoryJobCreationFailing**
   - Condition: Job creation errors > 0 for 5 minutes
   - Impact: Workflows cannot execute
   - Action: Check controller logs, RBAC permissions

2. **ObservatoryControllerDown**
   - Condition: No metrics or reconciliations for 5 minutes
   - Impact: Operator is not functioning
   - Action: Check pod status, restart controller

3. **ObservatoryWebhookFailures**
   - Condition: Webhook errors > 0.1/sec for 5 minutes
   - Impact: Invalid CRs may be created
   - Action: Check webhook configuration, certificates

### Warning Alerts (severity: warning)

4. **ObservatoryHighReconciliationLatency**
   - Condition: P99.9 latency > 60s for 10 minutes
   - SLO: 99.9% of reconciliations under 60s
   - Action: Investigate slow reconciliations, scale controller

5. **ObservatoryHighErrorRate**
   - Condition: Error rate > 1% for 5 minutes
   - SLO: <1% error rate
   - Action: Check controller logs, investigate failures

6. **ObservatoryHighRetryRate**
   - Condition: Retries > 0.5/sec for 10 minutes
   - Impact: Workflows are struggling to complete
   - Action: Check Job logs, resource constraints

7. **ObservatoryJobTimeouts**
   - Condition: Timeout rate > 5% for 10 minutes
   - Impact: Workflows taking too long
   - Action: Increase timeouts, optimize Jobs

### Info Alerts (severity: info)

8. **ObservatoryWorkflowSlowdown**
   - Condition: Current avg duration > 2x historical avg for 15 minutes
   - Impact: Performance degradation
   - Action: Investigate cluster performance, resource usage

## Grafana Dashboards

### Quick Dashboard Import

Import the dashboard JSON from the Prometheus UI or use this PromQL query template:

```promql
# Reconciliation Rate
rate(observatory_reconcile_total[5m])

# Error Rate
sum(rate(observatory_reconcile_errors_total[5m]))
/
sum(rate(observatory_reconcile_total[5m]))

# P99 Latency
histogram_quantile(0.99, rate(observatory_reconcile_duration_seconds_bucket[5m]))

# Active Workflows
observatory_active_workflows

# Job Success Rate
sum(rate(observatory_job_completed_total{status="success"}[5m]))
/
sum(rate(observatory_job_completed_total[5m]))
```

### Example Dashboard Panels

**Panel 1: Reconciliation Overview**

```json
{
  "title": "Reconciliation Rate",
  "targets": [
    {
      "expr": "rate(observatory_reconcile_total[5m])",
      "legendFormat": "{{result}}"
    }
  ]
}
```

**Panel 2: SLO Compliance**

```json
{
  "title": "Error Rate SLO (Target: <1%)",
  "targets": [
    {
      "expr": "sum(rate(observatory_reconcile_errors_total[5m])) / sum(rate(observatory_reconcile_total[5m])) * 100"
    }
  ],
  "alert": {
    "threshold": 1.0
  }
}
```

## Verification

### Check ServiceMonitor

```bash
kubectl get servicemonitor -n observatory-system
kubectl describe servicemonitor observatory-controller-metrics -n observatory-system
```

### Check Prometheus Targets

In Prometheus UI:

1. Go to Status → Targets
2. Look for `observatory-controller-metrics`
3. Should show "UP" status

### Test Alerting Rules

```bash
# Get PrometheusRule
kubectl get prometheusrule -n observatory-system

# Check rule evaluation in Prometheus UI
# Go to Alerts → Show all PrometheusRules
# Filter: observatory
```

### Query Metrics

```bash
# Port-forward to Prometheus
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090

# Open browser: http://localhost:9090
# Query: observatory_reconcile_total
```

## Troubleshooting

### No Metrics Appearing

1. **Check ServiceMonitor is created:**

   ```bash
   kubectl get servicemonitor -n observatory-system
   ```

2. **Check controller metrics endpoint:**

   ```bash
   kubectl port-forward -n observatory-system svc/observatory-controller-metrics 8080:8080
   curl http://localhost:8080/metrics
   ```

3. **Check Prometheus logs:**

   ```bash
   kubectl logs -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0
   ```

4. **Verify RBAC for Prometheus:**
   ```bash
   kubectl get clusterrole prometheus-kube-prometheus-prometheus
   ```

### Alerts Not Firing

1. **Check PrometheusRule syntax:**

   ```bash
   kubectl describe prometheusrule observatory-controller-alerts -n observatory-system
   ```

2. **Check Alertmanager:**

   ```bash
   kubectl get alertmanager -n monitoring
   ```

3. **Verify rule evaluation:**
   - In Prometheus UI: Status → Configuration
   - Look for `observatory-controller.rules`

### High Memory Usage

Reduce metric retention or cardinality:

```yaml
# In servicemonitor.yaml, add:
spec:
  endpoints:
    - metricRelabelings:
        - sourceLabels: [__name__]
          regex: "observatory_(less_important_metrics).*"
          action: drop
```

## Advanced Configuration

### Custom Scrape Interval

Edit `servicemonitor.yaml`:

```yaml
spec:
  endpoints:
    - interval: 15s # Scrape every 15s instead of 30s
      scrapeTimeout: 5s
```

### Label Remapping

Add labels to all metrics:

```yaml
spec:
  endpoints:
    - relabelings:
        - sourceLabels: [__meta_kubernetes_pod_node_name]
          targetLabel: node
```

### Multiple Prometheus Instances

If running multiple Prometheus instances, use labels:

```yaml
metadata:
  labels:
    prometheus: primary # Only scraped by 'primary' Prometheus
```

## Integration with Alertmanager

Configure Alertmanager routes for Observatory alerts:

```yaml
# alertmanager-config.yaml
route:
  receiver: default
  routes:
    - match:
        prometheus: observatory
      receiver: observatory-oncall
      group_by: [alertname, cluster]
      group_wait: 10s
      group_interval: 5m
      repeat_interval: 3h

receivers:
  - name: observatory-oncall
    pagerduty_configs:
      - service_key: <your-key>
        severity: "{{ .GroupLabels.severity }}"
```

## References

- [Prometheus Operator Documentation](https://prometheus-operator.dev/)
- [ServiceMonitor Spec](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitor)
- [PrometheusRule Spec](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PrometheusRule)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
