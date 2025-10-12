#!/usr/bin/env python3
import json, csv, pathlib
from datetime import datetime, timezone
import os

ROOT = pathlib.Path(__file__).resolve().parents[1]
OUT = ROOT / "out"
OUT.mkdir(parents=True, exist_ok=True)

def znow():
    sde = int(os.environ.get("SOURCE_DATE_EPOCH", "0"))
    t = datetime.fromtimestamp(sde, tz=timezone.utc) if sde else datetime.now(timezone.utc)
    return t.replace(microsecond=0).isoformat().replace("+00:00", "Z")

def write_verify_stats_csv():
    p = OUT / "verify_stats.csv"
    rows = [
        # timestamp, certs, ok, fails, elapsed
        [znow(), 10, 10, 0, 12],
        [znow(), 12, 12, 0, 11],
    ]
    with p.open("w", newline="", encoding="utf-8") as fh:
        w = csv.writer(fh, lineterminator="\n")
        w.writerow(["timestamp","certs","ok","fails","elapsed"])
        w.writerows(rows)

def write_verify_stats_json():
    # Schema requires: timestamp, certs, ok, fails, elapsed
    p = OUT / "verify_stats.sample.json"
    obj = {
        "timestamp": znow(),
        "certs": 12,
        "ok": 12,
        "fails": 0,
        "elapsed": 11
    }
    p.write_text(json.dumps(obj, indent=2), encoding="utf-8")

def write_anomalies_jsonl():
    p = OUT / "anomalies.jsonl"
    lines = [
        {"timestamp": znow(), "elapsed": 10, "factor": 1.2},
        {"timestamp": znow(), "elapsed": 11, "factor": 1.0},
    ]
    p.write_text("\n".join(json.dumps(x, separators=(",", ":")) for x in lines) + "\n", encoding="utf-8")

if __name__ == "__main__":
    write_verify_stats_csv()
    write_verify_stats_json()
    write_anomalies_jsonl()
    print("OK: wrote sample artifacts to out/")
