# Operations Runbook (Starter)
- Monitor controller logs: `kubectl logs -n observatory-system -l control-plane=controller-manager -f`
- Check workflows: `kubectl get observatoryruns`
- Check Jobs: `kubectl get jobs -l obs.seventh/run=<name>`
- Common issues: RBAC, cert-manager not installed, webhook CA injection missing.
