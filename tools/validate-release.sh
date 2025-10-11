
#!/usr/bin/env bash
set -euo pipefail; export LC_ALL=C LANG=C TZ=UTC
grep -q "## " CHANGELOG.md
test -s CHANGELOG.json
echo "OK: release notes present"
