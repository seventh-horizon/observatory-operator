
# Meta Observability Contract

This document pinches the runtime telemetry invariants for long-term stability.

## verify_stats*.json
- `timestamp`: ISO8601 UTC with trailing `Z`
- `certs`, `ok`, `fails`, `elapsed`: integers (seconds)
- Optional: `req`

## CSV Row
Header (exact): `timestamp,certs,ok,fails,elapsed`

## anomalies.jsonl lines
Each line:
```json
{"timestamp":"...Z","elapsed":14,"median":12,"factor":1.2}
```

## Dashboard `/data.json` shape (minimum)
```json
{
  "generated_at":"...Z",
  "total": 6,
  "ok": 6,
  "fails": 0,
  "timestamps": [...],
  "elapsed": [...],
  "ok_per_run": [...],
  "fails_per_run": [...],
  "anomalies": [...],
  "alert_factor": 2.0,
  "window": 5
}
```

CI runs on Ubuntu and macOS and must pass without network access.
