#!/usr/bin/env bash
# git diff textconv driver for tofu/altair/terraform.tfstate — decrypts the
# blob git hands us and pretty-prints it so history stays readable, without
# ever touching the real state file. Registered via tfstate-diff-setup.sh.
set -euo pipefail

BLOB="$1"
ROOT="$(git rev-parse --show-toplevel)"
CACHE="$ROOT/tofu/altair/.tfstate-decrypt-cache"

if [ ! -f "$CACHE/.setup-complete" ]; then
  echo "(tfstate diff not set up — run tofu/altair/tfstate-diff-setup.sh once)" >&2
  cat "$BLOB"
  exit 0
fi

cp "$BLOB" "$CACHE/terraform.tfstate"
export TF_VAR_tofu_encryption_passphrase
TF_VAR_tofu_encryption_passphrase="$(sops -d --extract '["passphrase"]' "$ROOT/tofu/altair/encryption-passphrase.sops.yaml" 2>/dev/null)"

if [ -z "$TF_VAR_tofu_encryption_passphrase" ]; then
  echo "(could not decrypt encryption-passphrase.sops.yaml)" >&2
  exit 0
fi

if ! ( cd "$CACHE" && tofu state pull 2>/dev/null | jq . ); then
  echo "(unable to decrypt this revision — empty, invalid, or pre-encryption state)" >&2
fi
