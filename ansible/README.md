# Altair — Ansible

Covers steps 3–5 of `bootstrap/altair.md`: installing k3s inside the LXC,
the Proxmox-LXC compatibility fixes, removing the cloud-provider taint, and
creating the `sops-age` secret. Step 2 (creating the LXC itself) is
`../tofu/altair/`. Flux bootstrap (step 6) and the AdGuard rewrite (step 7)
stay manual/one-off — see the README there for why.

## Source of truth

Don't hand-edit `/etc/rancher/k3s/config.yaml`, `resolv.conf`, the taint, or
the `sops-age` secret on Altair — change the role and re-run the playbook.

## Roles

- `k3s_lxc` — installs k3s, writes `config.yaml`/`resolv.conf`, the
  `/proc/sys` remount systemd override, and waits for the node to go Ready.
  Runs via `pct exec` on the Proxmox host, not direct SSH to the container —
  there's no SSH key authorized inside the LXC itself, matching how the
  bootstrap doc only ever uses `pct enter`/`pct exec`.
- `k3s_cluster_init` — runs locally against the `altair` kubectl context:
  removes the cloud-provider taint and creates the `sops-age` secret.

## Usage

```bash
cd ansible
ansible-galaxy collection install -r requirements.yml
ansible-playbook playbook.yml
```

Requires:
- SSH root access to the Proxmox host (`10.0.0.166`) — same access already
  used for manual maintenance. The `community.proxmox.proxmox_pct_remote`
  connection plugin (from `requirements.yml`) uses this single hop and runs
  everything inside the container via `pct exec`.
- `kubectl` on the control machine with an `altair` context configured,
  for the taint-removal and secret-creation tasks (they run locally, not
  on the LXC).
- The age key at `~/.config/sops/age/keys.txt` (macOS: `~/Library/Application
  Support/sops/age/keys.txt` — override with `-e sops_age_key_path=...`).
  Pull it from 1Password first if it's not already on disk.

Safe to re-run: every task is idempotent (checks the current state of the
config file / taint / secret before changing anything).
