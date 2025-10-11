
#!/usr/bin/env bash
set -euo pipefail; export LC_ALL=C LANG=C TZ=UTC
: "${TAG:=v0.0.0}"
test -f CHANGELOG.md || echo -e "# Changelog

## ${TAG}
- Initial." > CHANGELOG.md
jq -n --arg tag "$TAG" '{tag:$tag, notes:"placeholder"}' > CHANGELOG.json
# placeholder: add sigs in real pipeline
