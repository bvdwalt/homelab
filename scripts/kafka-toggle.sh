#!/usr/bin/env bash
# Turn the Kafka + order-pipeline learning stack off/on to free ~2 GiB of memory.
# off: saves the cluster ID, suspends Flux, scales services to 0, deletes the Kafka CR and node pool (PVC and data are kept), then stops the Strimzi operator.
# on:  starts the operator, resumes Flux, restores the cluster ID, then scales the services back up.
set -euo pipefail

CTX="--context=altair"
K="kubectl $CTX -n kafka"

case "${1:-}" in
  off)
    ID=$($K get kafka order-pipeline -o jsonpath='{.status.clusterId}')
    [ -n "$ID" ] || { echo "no clusterId on Kafka CR, aborting" >&2; exit 1; }
    # a recreated Kafka CR gets a new ID that no longer matches the PVC
    $K create configmap kafka-cluster-id --from-literal=id="$ID" --dry-run=client -o yaml | $K apply -f -
    flux $CTX suspend kustomization order-pipeline kafka
    kubectl $CTX -n order-pipeline scale deployment --all --replicas=0
    $K delete kafka order-pipeline
    $K delete kafkanodepool dual-role
    # the operator must stay up until the broker is gone
    $K wait --for=delete pod/order-pipeline-dual-role-0 --timeout=120s
    $K scale deployment strimzi-cluster-operator --replicas=0
    ;;
  on)
    ID=$($K get configmap kafka-cluster-id -o jsonpath='{.data.id}')
    [ -n "$ID" ] || { echo "no saved cluster ID, aborting" >&2; exit 1; }
    # flux resume would block on the health check, which fails until the ID is restored
    kubectl $CTX -n flux-system patch kustomization kafka --type=merge -p '{"spec":{"suspend":false}}'
    $K scale deployment strimzi-cluster-operator --replicas=1
    until $K get kafka order-pipeline >/dev/null 2>&1; do sleep 2; done
    # the operator may write its own ID first, so keep re-applying until it sticks
    until [ "$($K get configmap order-pipeline-dual-role-0 -o jsonpath='{.data.cluster\.id}' 2>/dev/null)" = "$ID" ]; do
      $K patch kafka order-pipeline --subresource=status --type=merge -p "{\"status\":{\"clusterId\":\"$ID\"}}" >/dev/null || true
      sleep 3
    done
    $K delete pod order-pipeline-dual-role-0 --ignore-not-found
    $K wait --for=condition=Ready pod/order-pipeline-dual-role-0 --timeout=300s
    flux $CTX resume kustomization order-pipeline
    # scaling to 0 leaves replicas owned by kubectl, so Flux does not restore it
    kubectl $CTX -n order-pipeline scale deployment --all --replicas=1
    kubectl $CTX -n order-pipeline scale deployment inventory-service --replicas=2
    ;;
  *)
    echo "usage: $0 off|on" >&2
    exit 1
    ;;
esac
