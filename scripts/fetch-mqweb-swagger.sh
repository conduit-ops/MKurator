#!/usr/bin/env bash
# Fetch the IBM MQ mqweb Swagger 2.0 document from a running queue manager.
#
# Prerequisites:
#   - mqweb running with apiDiscovery-1.0 enabled (see docs/IBM_MQ_REST_API.md)
#   - MQWebUser/MQWebAdmin credentials with access to /ibm/api/docs
#
# Usage:
#   ./scripts/fetch-mqweb-swagger.sh [base-url] [output-file]
#
# Examples:
#   ./scripts/fetch-mqweb-swagger.sh
#   ./scripts/fetch-mqweb-swagger.sh https://localhost:9443 docs/schemas/mqweb-swagger-9.4.json
#   MQWEB_USER=admin MQWEB_PASS=secret ./scripts/fetch-mqweb-swagger.sh https://qm.example.com:9443

set -euo pipefail

BASE_URL="${1:-https://localhost:9443}"
OUT="${2:-docs/schemas/mqweb-swagger.json}"
USER="${MQWEB_USER:-}"
PASS="${MQWEB_PASS:-}"
CURL_EXTRA="${MQWEB_CURL_EXTRA:--k}"

mkdir -p "$(dirname "$OUT")"

auth=()
if [[ -n "$USER" ]]; then
  auth=(-u "${USER}:${PASS}")
fi

echo "Fetching Swagger from ${BASE_URL}/ibm/api/docs -> ${OUT}" >&2
curl -fsS ${CURL_EXTRA} "${auth[@]}" \
  -H "Accept: application/json" \
  "${BASE_URL}/ibm/api/docs" \
  -o "${OUT}"

# Basic validation
if ! python3 -c "import json; json.load(open('${OUT}'))" 2>/dev/null; then
  echo "ERROR: ${OUT} is not valid JSON. Check credentials and apiDiscovery feature." >&2
  exit 1
fi

echo "OK: $(wc -c < "${OUT}") bytes written to ${OUT}" >&2
echo "Browse interactively: ${BASE_URL}/ibm/api/explorer" >&2
