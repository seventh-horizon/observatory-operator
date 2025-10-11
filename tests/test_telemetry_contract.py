
import json, pathlib
from jsonschema import validate, Draft202012Validator

SCHEMA = json.loads((pathlib.Path('schemas')/'verify_stats.schema.json').read_text())

def test_verify_stats_json_shape():
    samples = sorted(pathlib.Path('out').glob('verify_stats*.json'))
    assert samples, "no verify_stats*.json samples found"
    for p in samples:
        data = json.loads(p.read_text())
        Draft202012Validator(SCHEMA).validate(data)

def test_csv_header_and_rows():
    import csv
    p = pathlib.Path('out/verify_stats.csv')
    assert p.exists(), "missing CSV sample"
    with p.open(newline='') as fh:
        r = csv.reader(fh)
        header = next(r)
        assert header == ["timestamp","certs","ok","fails","elapsed"]
        for row in r:
            assert len(row)==5
            # light type checks
            assert row[0].endswith("Z")
            for x in row[1:]:
                assert x.isdigit()
