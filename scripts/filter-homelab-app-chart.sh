#!/usr/bin/env bash
# Keeps only manifests rendered from our own homelab-app chart.
set -euo pipefail

in_file="$1"
out_file="$2"

awk '
  /^---$/ { if (keep) printf "%s---\n", doc; doc=""; keep=0; next }
  { doc = doc $0 "\n"; if ($0 ~ /^# Source: homelab-app\//) keep=1 }
  END { if (keep) printf "%s", doc }
' "$in_file" > "$out_file"
