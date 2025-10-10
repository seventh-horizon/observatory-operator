# release-hardened

**Seventh Horizon**’s reproducible release system.  
Every artifact is byte-identical across platforms, cryptographically signed, and timeline-indexed for long-term auditability.

Core features:
- Deterministic tar/gzip packer with Python mtime=0 + zlib version capture
- Signed release manifests and changelogs (`.asc`, `.sha256`)
- Temporal topology tracking (`Φ`, `κ`, `φ` fields with MDS embedding)
- Timeline index validator and visualization
- Provenance pipeline (OIDC, cosign, minisign)
- Cross-platform reproducibility verified in CI

> “Build once, verify forever.”
