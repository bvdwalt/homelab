# Altair LXC (OpenTofu)

Declares the Proxmox LXC container (`pct create` equivalent) hosting Altair's
k3s node; covers step 2 of `bootstrap/altair.md`. Imported from the live
container, so `main.tf`'s values (node `pve`, 12288MB memory, `cheetah` ZFS
at 96G, 16 cores) reflect real state, not bootstrap.md's smaller originals.
Re-verify against `pct config <vmid>` before applying if time has passed:
cores, disk size, and bind mounts have drifted from this file before.

## Source of truth

Don't hand-edit `/etc/pve/lxc/100.conf` or run `pct set`/`pct resize` on
Altair. Change `main.tf`/`variables.tf` and apply instead.

## What this does NOT cover

Proxmox's API has no fields for `lxc.apparmor.profile`, `lxc.cgroup2.devices.allow`,
`lxc.mount.entry`, `lxc.cap.drop`, or `lxc.seccomp.profile`: they only exist as raw
lines in `/etc/pve/lxc/<vmid>.conf`, and no Proxmox provider sets them natively.
A `null_resource` with a `remote-exec` SSH provisioner (to the Proxmox host, not
the container) appends those lines and restarts, the same as the bootstrap
doc's manual `echo >> 100.conf` step plus everything added since (GPU render
node, tun device, three media bind mounts). It also `mkdir -p`s each
`cheetah_bind_mounts` source, so a new bind mount is just a list entry, no
manual directory step.

Everything from "Install k3s" onward (step 3+) is in `../../ansible/`, not here.

## Secrets: SOPS-encrypted, not tfvars

`terraform.tfvars` is committed and holds no secrets: `template_file` and
`ssh_public_keys` aren't sensitive. The two things that are, the Proxmox API
token and the state/plan encryption passphrase, live in their own
SOPS-encrypted YAML files (same age key as the rest of this repo, see
`.sops.yaml`) and get exported as `TF_VAR_*` env vars at runtime instead:

```bash
export TF_VAR_tofu_encryption_passphrase=$(sops -d --extract '["passphrase"]' encryption-passphrase.sops.yaml)
export TF_VAR_proxmox_api_token=$(sops -d --extract '["api_token"]' proxmox-token.sops.yaml)
```

`proxmox-token.sops.yaml` doesn't exist yet: the token used to validate this
config was revoked after (see repo history). Create a new one when you're
ready to apply:

```bash
ssh root@10.0.0.166 "pveum user token add root@pam tofu --privsep 0"
```

Then SOPS-encrypt the result yourself, e.g.:

```bash
printf 'api_token: "root@pam!tofu=<uuid>"\n' > proxmox-token.sops.yaml
sops -e -i proxmox-token.sops.yaml
```

Revoke it again with `pveum user token remove root@pam tofu` when done, to
avoid a standing credential on the host.

## State and plan encryption

`terraform.tfstate` is encrypted at rest using OpenTofu's native
[state encryption](https://opentofu.org/docs/language/state/encryption/)
(`versions.tf`'s `encryption` block): AES-GCM with a key derived via PBKDF2
from `tofu_encryption_passphrase`. It's genuine ciphertext (an
`encrypted_data` blob, no resource attributes in the clear), so it's safe to
commit, hence no longer gitignored. OpenTofu encrypts/decrypts it
transparently; you never handle plaintext state as long as the passphrase
env var is set.

Re-bootstrapping (e.g. a fresh `tofu import` producing plaintext state)
migrates the same way this one did:

```bash
export TF_VAR_tofu_encryption_passphrase=$(sops -d --extract '["passphrase"]' encryption-passphrase.sops.yaml)
tofu state pull > /tmp/state.json   # reads plaintext via a temporary unencrypted fallback method
tofu state push /tmp/state.json     # rewrites it through the real aes_gcm method
rm /tmp/state.json
```

(Add a temporary `fallback { method = method.unencrypted.migrate }` under
`state`/`plan` in `versions.tf` for that one round-trip, then remove it.)

### Readable diffs

Encrypting the whole state as one AES-GCM blob means `git diff`/`git log -p`
show nothing useful by default: every write produces a different ciphertext
string with no hint of which field changed. To get real diffs back without
weakening at-rest security, run this once per machine/clone:

```bash
./tfstate-diff-setup.sh
```

This is a *local* git config (`git config diff.tfstate-decrypt.textconv`), not
something `.gitattributes` can carry: git keeps executable diff commands out
of versioned config for security, so every clone needs to run the setup
script once. After that, `git diff`/`git log -p`/`git show` on
`terraform.tfstate` transparently decrypt each revision (via a throwaway
harness in `.tfstate-decrypt-cache/`, gitignored) for display only; the
stored/pushed blob is never touched.

## Usage

```bash
cd tofu/altair
export TF_VAR_tofu_encryption_passphrase=$(sops -d --extract '["passphrase"]' encryption-passphrase.sops.yaml)
export TF_VAR_proxmox_api_token=$(sops -d --extract '["api_token"]' proxmox-token.sops.yaml)
tofu init
tofu plan
tofu apply
```

Requires:
- SSH root access to the Proxmox host from wherever you run `tofu apply`
  (same access as manual maintenance, see repo memory).
- The Debian 12 LXC template already downloaded on the host
  (`local:vztmpl/debian-12-standard_*.tar.zst`); get the exact filename with
  `pveam list local` and update `template_file` in `terraform.tfvars` if it
  changed.

The container resource has `lifecycle.ignore_changes` on `operating_system`
and `initialization[0].user_account`: both are creation-time-only fields that
force a destroy-and-recreate if OpenTofu diffs them against an already-
imported container. Don't remove that block without understanding why (see
git history for what happened the one time it wasn't there).

## Verify

```bash
ssh root@10.0.0.166 pct config 100
ssh root@10.0.0.166 pct status 100
```
