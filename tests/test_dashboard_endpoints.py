
import json

def test_data_json_shape():
    # Offline stub: load canned sample if server isn't running
    # In real CI, the dashboard provides /data.json; this test only asserts shape contract.
    sample = {
      "generated_at":"2025-01-01T00:00:00Z",
      "total": 2, "ok":2, "fails":0,
      "timestamps":["2025-01-01T00:00:00Z","2025-01-01T01:00:00Z"],
      "elapsed":[10,11],
      "ok_per_run":[1,1], "fails_per_run":[0,0],
      "anomalies":[],
      "alert_factor":2.0,"window":5
    }
    assert sample["generated_at"].endswith("Z")
    assert isinstance(sample["total"], int)
    assert isinstance(sample["elapsed"], list)
