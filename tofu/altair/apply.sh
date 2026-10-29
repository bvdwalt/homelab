#!/usr/bin/env bash
# Applies the altair config and stages the resulting encrypted state, so the
# state ships in the same PR as the change that produced it.
set -euo pipefail
cd "$(dirname "$0")"

if [ "$(git branch --show-current)" = main ]; then
  echo "Run this from a feature branch or worktree, main is protected." >&2
  exit 1
fi

if ! git diff --quiet -- terraform.tfstate; then
  echo "terraform.tfstate has uncommitted changes from an earlier run. Commit or discard them first." >&2
  exit 1
fi

export TF_VAR_tofu_encryption_passphrase
TF_VAR_tofu_encryption_passphrase=$(sops -d --extract '["passphrase"]' encryption-passphrase.sops.yaml)
export TF_VAR_proxmox_api_token
TF_VAR_proxmox_api_token=$(sops -d --extract '["api_token"]' proxmox-token.sops.yaml)

tofu init -input=false

# Stage state even if apply fails: a partial apply still changes it.
rc=0
tofu apply "$@" || rc=$?

git add terraform.tfstate .terraform.lock.hcl
echo "State staged. Commit it with your change before pushing."
exit "$rc"
