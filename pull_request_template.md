## Summary
<!-- What does this PR change? Why? -->

## Type
- [ ] Feature
- [ ] Bug fix
- [ ] Docs
- [ ] Chore / Tech debt

## Linked Work
<!-- Use "Fixes #123" (auto-close) or "Refs #123" (cross-reference) -->
Fixes #

## Release & Versioning
- [ ] Updates `CHANGELOG.md`
- [ ] Keeps `VERSION` accurate for this branch
- [ ] Backwards compatible (or breaking changes documented)

## Validations Done
- [ ] 1/4: CRD renders cleanly (`kustomize build config/default`)
- [ ] 2/4: Webhook healthy (svc + endpoints on :9443)
- [ ] 3/4: Status propagation (sample: `ok-minimal.yaml`)
- [ ] 4/4: Failure policy / retries behavior (samples)

## How to Test
```bash
# Kind + cert-manager quickstart
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
kustomize build config/default | kubectl apply -f -

# Happy path
kubectl apply -f config/samples/ok-minimal.yaml
kubectl -n observatory-system get observatoryrun ok-minimal -o yaml | yq '.status'

# Failure policy (Stop)
kubectl apply -f config/samples/with-failure-policy.yaml
kubectl -n observatory-system get observatoryrun fp-stop-demo -o yaml | yq '.status'

# Retries
kubectl apply -f config/samples/with-retries.yaml
kubectl -n observatory-system get observatoryrun retry-demo -o yaml | yq '.status'

<!-- Paste controller logs or screenshots showing behavior -->
