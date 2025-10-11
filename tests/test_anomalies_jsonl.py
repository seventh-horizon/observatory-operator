
import json, pathlib

def test_anomalies_jsonl_shape():
    p = pathlib.Path('out/anomalies.jsonl')
    assert p.exists(), "missing anomalies.jsonl sample"
    lines = [l for l in p.read_text().splitlines() if l.strip()]
    assert lines, "no lines in anomalies.jsonl"
    for line in lines:
        obj = json.loads(line)
        assert set(obj.keys()) >= {"timestamp","elapsed","factor"}
        assert obj["timestamp"].endswith("Z")
        assert isinstance(obj["elapsed"], int)
        assert isinstance(obj["factor"], (int,float))
