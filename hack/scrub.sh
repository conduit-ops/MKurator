#!/usr/bin/env bash
# Fail if staged files contain forbidden employer or legacy-project strings.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

files="$(git diff --cached --name-only 2>/dev/null | grep -Ev '^hack/scrub(-patterns\.txt|\.sh)$' || true)"
if [ -z "${files}" ]; then
  echo "scrub: no staged files (ok)"
  exit 0
fi

if echo "${files}" | xargs rg -il -f hack/scrub-patterns.txt 2>/dev/null; then
  echo "scrub: forbidden strings in staged files" >&2
  exit 1
fi

echo "scrub: ok"
