# homelab

Raspberry Pi 5 running k3s, with Traefik for ingress, cert-manager for TLS, and local-path for persistent storage.

## Architecture

```
TrueNAS SCALE (10.13.1.165)
└── Postgres — shared database for miniflux, linkwarden, atuin, metering-api

Raspberry Pi 5 (10.13.1.164) — k3s
├── Traefik         — ingress, *.greedo.net
├── cert-manager    — wildcard TLS via Let's Encrypt + Cloudflare DNS-01
└── services/       — see k8s/raspi/apps/
```

AdGuard DNS (running on TrueNAS) resolves `*.greedo.net → 10.13.1.164`.

## Repo layout

```
k8s/
  charts/
    homelab-app/       — shared Helm chart used by all services
  raspi/
    apps/              — HelmRelease resources for each service
    infrastructure/
      configs/         — CRD-dependent configs (ClusterIssuer, certs, TLS store)
      secrets/         — SOPS-encrypted Secrets + namespace pre-creation
      sources/         — HelmRepository sources
```

## Prerequisites

```bash
brew install helm kubectl sops fluxcd/tap/flux age
```

Your age private key must exist at `~/Library/Application Support/sops/age/keys.txt`.

## Bootstrap

### 1. Install k3s on the Pi

```bash
curl -sfL https://get.k3s.io | sh -
```

Copy `/etc/rancher/k3s/k3s.yaml` to `~/.kube/config` on your local machine and set the server IP to `10.13.1.164`.

### 2. Install the SOPS age key into the cluster

```bash
kubectl create namespace flux-system
kubectl create secret generic sops-age \
  --namespace=flux-system \
  --from-file=age.agekey="$HOME/Library/Application Support/sops/age/keys.txt"
```

### 3. Bootstrap Flux

Requires a fine-grained GitHub PAT scoped to this repo with Contents and Administration read/write.

```bash
flux bootstrap github \
  --owner=bvdwalt \
  --repository=homelab \
  --path=k8s/raspi \
  --personal
```

Flux reconciles every 10 minutes. Force an immediate sync with:

```bash
flux reconcile kustomization flux-system --with-source
```

## Secrets

Secrets are SOPS-encrypted with an age key. Edit a secret with:

```bash
sops k8s/raspi/infrastructure/secrets/<name>.sops.yaml
```

The `.sops.yaml` creation rule applies automatically to files under `k8s/raspi/infrastructure/secrets/`.
