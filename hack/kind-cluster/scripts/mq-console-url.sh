#!/usr/bin/env bash
# Print IBM MQ web console URL and credentials for the local kind cluster.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if terraform -chdir="${ROOT_DIR}/terraform" output -raw mq_web_console_url >/dev/null 2>&1; then
  terraform -chdir="${ROOT_DIR}/terraform" output -raw mq_web_console_url
else
  echo "https://mq.localhost:30443/ibmmq/console/"
fi

cat <<'EOF'

Login: admin / passw0rd (hack/kind-cluster terraform default; local dev only)
Open in a browser that trusts mkcert (*.localhost), or run: mkcert -install
EOF
