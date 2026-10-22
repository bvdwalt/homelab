#!/usr/bin/env bash
# Turn the Kafka + order-pipeline learning stack off/on to free ~2 GiB of memory.
# off: suspends Flux, scales services to 0, deletes the Kafka CR (PVC and data are kept).
# on:  resumes Flux, which recreates everything from git.
set -euo pipefail

CTX="--context=altair"

case "${1:-}" in
  off)
    flux $CTX suspend kustomization order-pipeline kafka
    kubectl $CTX -n order-pipeline scale deployment --all --replicas=0
    kubectl $CTX -n kafka delete kafka order-pipeline
    ;;
  on)
    flux $CTX resume kustomization kafka
    flux $CTX resume kustomization order-pipeline
    ;;
  *)
    echo "usage: $0 off|on" >&2
    exit 1
    ;;
esac
