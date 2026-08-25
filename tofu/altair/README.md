# Altair LXC — OpenTofu

Declares the Proxmox LXC container (`pct create` equivalent) that hosts Altair's
k3s node. Covers step 2 of `bootstrap/altair.md`. Imported from the real,
already-running container — `main.tf`'s values (node name `pve`, 12288MB
memory, `cheetah` ZFS storage at 96G, 16 cores, etc.) reflect verified live
state, not the smaller numbers in `bootstrap/altair.md`, which predate this
container's growth. Re-verify against `pct config <vmid>` before applying if
much time has passed — cores/disk size and the media bind mounts have drifted
from this file before without Tofu being run.

## What this does NOT cover

Proxmox's API has no fields for `lxc.apparmor.profile`, `lxc.cgroup2.devices.allow`,
`lxc.mount.entry`, `lxc.cap.drop`, or `lxc.seccomp.profile` — these only exist as
raw lines in `/etc/pve/lxc/<vmid>.conf`. No Terraform/OpenTofu Proxmox provider
can set them natively. This config creates the container, then uses a
`null_resource` with a `remote-exec` SSH provisioner (connecting to the Proxmox
host, not the container) to idempotently append those lines to the conf file
and restart the container — the same thing the bootstrap doc's `echo >> 100.conf`
step does by hand, plus everything added ad hoc since (GPU render node
renumbered to `card1`, a `tun` device for VPN, three media bind mounts).

Everything from "Install k3s" onward (step 3+) is in `../../ansible/`, not here.

## Secrets: SOPS-encrypted, not tfvars

`terraform.tfvars` is committed and contains no secrets — `template_file` and
`ssh_public_keys` aren't sensitive. The two things that are (the Proxmox API
token and the state/plan encryption passphrase, see below) are never written
to a `.tfvars` file at all. Each lives in its own SOPS-encrypted YAML file
(same age key as everywhere else in this repo — see `.sops.yaml`) and gets
exported as a `TF_VAR_*` env var at runtime instead:

```bash
export TF_VAR_tofu_encryption_passphrase=$(sops -d --extract '["passphrase"]' encryption-passphrase.sops.yaml)
export TF_VAR_proxmox_api_token=$(sops -d --extract '["api_token"]' proxmox-token.sops.yaml)
```

`proxmox-token.sops.yaml` doesn't exist yet — the token created to validate
this config was revoked after use (see repo history). Create a new one when
you're ready to actually apply:

```bash
ssh root@10.0.0.166 "pveum user token add root@pam tofu --privsep 0"
```

Then SOPS-encrypt the result yourself, e.g.:

```bash
printf 'api_token: "root@pam!tofu=<uuid>"\n' > proxmox-token.sops.yaml
sops -e -i proxmox-token.sops.yaml
```

Revoke it again with `pveum user token remove root@pam tofu` once you're done,
if you don't want a standing credential on the host between sessions.

## State and plan encryption

`terraform.tfstate` is encrypted at rest using OpenTofu's native
[state encryption](https://opentofu.org/docs/language/state/encryption/)
(`versions.tf`'s `encryption` block) — AES-GCM with a key derived via PBKDF2
from `tofu_encryption_passphrase`. The file is genuine ciphertext (an
`encrypted_data` blob, no resource attributes in the clear), so it's safe to
commit — that's why it's no longer gitignored. OpenTofu encrypts/decrypts it
transparently on every command; you never handle plaintext state yourself as
long as the passphrase env var is set.

If you ever need to re-bootstrap this (e.g. a fresh `tofu import` producing a
brand-new plaintext state), migrate it the same way this one was:

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
show nothing useful by default — every write produces a totally different
ciphertext string, with no indication of which field actually changed. To get
real diffs back without weakening at-rest security, run this once per
machine/clone:

```bash
./tfstate-diff-setup.sh
```

This is a *local* git config (`git config diff.tfstate-decrypt.textconv`),
not something `.gitattributes` can carry on its own — git deliberately keeps
executable diff commands out of versioned config for security, so every
clone that wants decrypted diffs needs to run the setup script once. After
that, `git diff`/`git log -p`/`git show` on `terraform.tfstate` transparently
decrypt each revision (via a throwaway harness in
`.tfstate-decrypt-cache/`, gitignored) for display only — the blob stored
and pushed is never touched.

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
- SSH root access to the Proxmox host from wherever you run `tofu apply` (same
  access already used for manual maintenance — see repo memory).
- The Debian 12 LXC template already downloaded on the host
  (`local:vztmpl/debian-12-standard_*.tar.zst`) — grab the exact filename with
  `pveam list local` and set `template_file` in `terraform.tfvars` if it's
  changed.

The container resource has `lifecycle.ignore_changes` on `operating_system`
and `initialization[0].user_account` — both are creation-time-only fields
that force a destroy-and-recreate if OpenTofu ever diffs them against an
already-imported container. Don't remove that block without understanding
why it's there (see git history for what happened the one time it wasn't).

## Verify

```bash
ssh root@10.0.0.166 pct config 100
ssh root@10.0.0.166 pct status 100
```
