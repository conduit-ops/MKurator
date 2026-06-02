#!/usr/bin/env bash
# Run runmqsc against the IBM MQ pod installed by hack/kind-cluster Terraform.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${STATE_DIR:-${ROOT_DIR}/.state}"
export KUBECONFIG="${KUBECONFIG:-${STATE_DIR}/kubeconfig.yaml}"

NS="${MQ_NAMESPACE:-ibm-mq}"
QMGR="${MQ_QMGR_NAME:-QM1}"
RELEASE="${MQ_HELM_RELEASE:-ibm-mq}"

pod_name() {
  local pod
  pod="$(kubectl get pods -n "${NS}" \
    -l "app.kubernetes.io/instance=${RELEASE}" \
    -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
  if [[ -z "${pod}" ]]; then
    pod="$(kubectl get pods -n "${NS}" \
      -o jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}' 2>/dev/null | awk '{print $1}')"
  fi
  if [[ -z "${pod}" ]]; then
    echo "no Running pod in namespace ${NS}; run: task cluster:up" >&2
    exit 1
  fi
  echo "${pod}"
}

POD="$(pod_name)"

if [[ $# -eq 0 ]]; then
  echo "Interactive runmqsc on ${POD} (queue manager ${QMGR}). Exit with ^D or 'end'." >&2
  exec kubectl exec -it -n "${NS}" "${POD}" -- runmqsc "${QMGR}"
fi

# One-shot: join arguments into a single MQSC line/command.
kubectl exec -i -n "${NS}" "${POD}" -- runmqsc "${QMGR}" <<<"$*"
