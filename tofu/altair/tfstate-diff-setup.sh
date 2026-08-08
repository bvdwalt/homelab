#!/usr/bin/env bash
# One-time per-machine setup for readable `git diff`/`git log -p` on the
# encrypted tofu/altair/terraform.tfstate. Registers a local git config
# textconv driver — this is NOT stored in the repo (git deliberately keeps
# textconv commands out of versioned .gitattributes), so run this once on
# every machine/clone that needs decrypted diffs.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
CACHE="$ROOT/tofu/altair/.tfstate-decrypt-cache"

mkdir -p "$CACHE"
cp "$ROOT/tofu/altair/tfstate-decrypt-harness/main.tf" "$CACHE/main.tf"

echo "Initializing decrypt harness (no network access needed, no providers)..."
export TF_VAR_tofu_encryption_passphrase
TF_VAR_tofu_encryption_passphrase="$(sops -d --extract '["passphrase"]' "$ROOT/tofu/altair/encryption-passphrase.sops.yaml")"
(cd "$CACHE" && tofu init -input=false >/dev/null)
touch "$CACHE/.setup-complete"

git config diff.tfstate-decrypt.textconv "$ROOT/tofu/altair/tfstate-diff.sh"

echo "Done. 'git diff' / 'git log -p' on tofu/altair/terraform.tfstate now show decrypted content."
