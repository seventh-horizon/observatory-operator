# release-hardened

Reproducible, verifiable, **self‑attesting** release pipeline.
Built to be auditor-calm across Linux and macOS.

## Highlights

- Deterministic bundles (same SHA256 across OSes)
- Minisign installer verified via Sigstore (OIDC, Rekor)
- Strict verifier (trusted-comment binding; key-ID normalization)
- CLI + Nox sessions + CI workflows ready to go

## Quick start

```bash
# local fast checks
make prepush

# build a deterministic release bundle (dry-run skeleton)
make bundle
```

See `docs/AUDITABILITY.md` for long-term verification guidance.
