# homelab

Single-node Kubernetes cluster (Thalos) running on a TrueNAS SCALE VM, with Traefik for
ingress, cert-manager for TLS, and democratic-csi for NFS-backed persistent storage.

## Architecture

```
TrueNAS SCALE (10.13.1.165)
└── Thalos VM (10.13.1.166) — Talos Linux, 4 vCPU, 8 GB
    ├── Traefik         — ingress, MetalLB VIP :80/:443, *.greedo.net
    ├── cert-manager    — wildcard TLS via Let's Encrypt + Cloudflare DNS-01
    ├── MetalLB         — L2 LoadBalancer, VIP 10.13.1.160
    ├── democratic-csi  — NFS StorageClass backed by Cheetah/k8s-nfs
    └── services/
        ├── vaultwarden
        ├── it-tools
        └── whoami
```

AdGuard DNS resolves `*.greedo.net → 10.13.1.160` (MetalLB VIP).

## Repo layout

```
ansible/               — provisions the Thalos VM and bootstraps the cluster
k8s/
  talos/               — talhelper config + SOPS-encrypted cluster secrets
  flux/                — Flux GitOps entrypoint (bootstrap path: k8s/flux)
  infrastructure/
    namespaces/        — privileged namespaces (democratic-csi, metallb-system)
    sources/           — HelmRepository sources
    releases/          — HelmRelease resources (managed by Flux, updated by Renovate)
    configs/           — CRD-dependent configs (MetalLB pools, ClusterIssuer, certs)
    secrets/           — SOPS-encrypted Secrets (Cloudflare token, TrueNAS credentials)
  manifests/           — plain Kubernetes manifests for stateless apps
```

## Prerequisites

```bash
brew install ansible talosctl talhelper helm kubectl sops fluxcd/tap/flux
```

```bash
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8
export TRUENAS_API_KEY="your-key-here"
export CLOUDFLARE_API_TOKEN="your-token-here"   # DNS:Edit for greedo.net zone
```

Your age key must exist at `~/Library/Application Support/sops/age/keys.txt`.
It is used to decrypt SOPS-encrypted secrets during bootstrap.

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

## GitOps with Flux

After the cluster is up, Flux takes over and keeps the cluster in sync with this repo.
Renovate runs as a GitHub App and opens PRs to bump Helm chart versions and container
image tags. Merging the PR is the only manual step — Flux applies it automatically.

### One-time Flux bootstrap

```bash
# 1. Populate SOPS secrets with real values (see "Secrets" section below)

# 2. Store your age private key in the cluster so Flux can decrypt SOPS secrets
kubectl create secret generic sops-age \
  --namespace=flux-system \
  --from-file=age.agekey="$HOME/Library/Application Support/sops/age/keys.txt" \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Bootstrap Flux — installs controllers and creates the GitRepository + sync loop
flux bootstrap github \
  --owner=bvdwalt \
  --repository=homelab \
  --path=k8s/flux \
  --personal
```

Flux reconciles every 10 minutes. Force an immediate sync with:
```bash
flux reconcile kustomization flux-system --with-source --kubeconfig k8s/talos/kubeconfig
```

### Secrets

The encrypted files in `k8s/infrastructure/secrets/` contain placeholder values.
Edit and re-encrypt them with your real credentials before bootstrapping Flux:

```bash
# Cloudflare API token (needs DNS:Edit permission for the greedo.net zone)
sops k8s/infrastructure/secrets/cloudflare.sops.yaml

# TrueNAS host, API key, and ZFS dataset paths for democratic-csi
sops k8s/infrastructure/secrets/democratic-csi.sops.yaml
```

### Helm chart versions

Chart versions in `k8s/infrastructure/releases/` are pinned. Renovate opens a PR
when a new version is available — merge it and Flux upgrades the release automatically.

To check what the cluster currently has installed:
```bash
helm list -A --kubeconfig k8s/talos/kubeconfig
```

If the pinned versions differ from what is currently running, update them to match
before bootstrapping Flux to avoid unintended upgrades.

### Renovate

Install the [Renovate GitHub App](https://github.com/apps/renovate) on this repository.
Configuration is in `renovate.json` — it watches:
- `k8s/infrastructure/releases/**` for Helm chart updates (`flux` manager)
- `k8s/manifests/**` for container image tag updates (`kubernetes` manager)
