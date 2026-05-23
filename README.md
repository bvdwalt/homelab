# homelab

Single-node Kubernetes cluster (Thalos) running on a TrueNAS SCALE VM, with Traefik for
ingress, cert-manager for TLS, and democratic-csi for NFS-backed persistent storage.

## Architecture

```
TrueNAS SCALE (10.13.1.165)
└── Thalos VM (10.13.1.166) — Talos Linux, 4 vCPU, 8 GB
    ├── Traefik         — ingress, hostPort :80/:443, *.greedo.net
    ├── cert-manager    — wildcard TLS via Let's Encrypt + Cloudflare DNS-01
    ├── democratic-csi  — NFS StorageClass backed by Cheetah/k8s-nfs
    └── services/
        ├── it-tools
        └── whoami
```

AdGuard DNS resolves `*.greedo.net → 10.13.1.166` for migrated services.

## Repo layout

```
ansible/      — provisions the Thalos VM and bootstraps the cluster
k8s/
  talos/      — talhelper config + SOPS-encrypted cluster secrets
  manifests/  — Helm values and Kubernetes manifests
```

## Prerequisites

```bash
brew install ansible talosctl talhelper helm kubectl sops
```

```bash
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8
export TRUENAS_API_KEY="your-key-here"
export CLOUDFLARE_API_TOKEN="your-token-here"   # DNS:Edit for greedo.net zone
```

Your age key must exist at `~/Library/Application Support/sops/age/keys.txt`.
It is used to decrypt `k8s/talos/talsecret.sops.yaml` during bootstrap.

## Usage

```bash
# 1. Provision the VM (boots Talos into maintenance mode)
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-vm.yml

# 2. Bootstrap the cluster — installs cert-manager, democratic-csi, Traefik, and services
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-bootstrap.yml

# Tear everything down
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-teardown.yml
```

Both playbooks are idempotent and safe to re-run. The teardown keeps
`k8s/talos/talsecret.sops.yaml` so the same cluster CA is reused on the next
bootstrap — delete it manually for a completely fresh cluster identity.
