#!/usr/bin/env bash
# Wait until mqweb is reachable on the Docker Compose port (direct TLS, no ingress Host).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=/dev/null
source "${ROOT}/hack/mq-docker/defaults.env" 2>/dev/null || true

ENDPOINT="${KURATOR_INTEGRATION_MQ_ENDPOINT:-https://127.0.0.1:9443}"
QMGR="${KURATOR_INTEGRATION_MQ_QMGR:-QM1}"
USER="${KURATOR_INTEGRATION_MQ_USER:-admin}"
PASS="${KURATOR_INTEGRATION_MQ_PASSWORD:-passw0rd}"
MAX_ATTEMPTS="${MQWEB_WAIT_ATTEMPTS:-60}"
SLEEP_SECS="${MQWEB_WAIT_SLEEP:-10}"

URL="${ENDPOINT%/}/ibmmq/rest/v3/admin/qmgr/${QMGR}"

echo "Waiting for mqweb at ${URL} …"

for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
  code="$(curl -sk -o /dev/null -w "%{http_code}" -u "${USER}:${PASS}" "${URL}" || true)"
  if [[ "${code}" =~ ^(200|401|403)$ ]]; then
    echo "mqweb ready (HTTP ${code}) after ${attempt} attempt(s)."
    exit 0
  fi
  echo "attempt ${attempt}/${MAX_ATTEMPTS}: HTTP ${code:-000}; retrying in ${SLEEP_SECS}s …"
  sleep "${SLEEP_SECS}"
done

echo "mqweb did not become ready in time" >&2
exit 1
