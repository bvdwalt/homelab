# homelab

![Validate](https://github.com/bvdwalt/homelab/actions/workflows/validate.yaml/badge.svg)
![TruffleHog](https://github.com/bvdwalt/homelab/actions/workflows/trufflehog.yaml/badge.svg)
![Checks](https://img.shields.io/github/checks-status/bvdwalt/homelab/main)

Two k3s clusters managed with Flux, running on a Raspberry Pi 5 and an HP Elite Mini 800 G9.

## Key components

| Component | Role |
|-----------|------|
| [Flux](https://fluxcd.io) | GitOps — all cluster state is reconciled from this repo |
| [Cilium](https://cilium.io) | CNI on both clusters, native routing on Raspi |
| [Traefik](https://traefik.io) | Ingress controller, TLS termination |
| [cert-manager](https://cert-manager.io) | Wildcard TLS certificates via Let's Encrypt + Cloudflare DNS-01 |
| [CNPG](https://cloudnative-pg.io) | CloudNativePG — shared PostgreSQL cluster on Altair |
| [AdGuard Home](https://adguard.com/adguard-home) | Local DNS with per-service rewrite rules |
| [Renovate](https://docs.renovatebot.com) | Automated dependency updates for Helm charts and container images |
| [SOPS + age](https://github.com/getsops/sops) | Secrets encrypted at rest in the repo |

## Architecture

```
Altair — HP Elite Mini 800 G9
├── Proxmox VE  (10.13.1.166)
└── k3s LXC     (10.13.1.167)  — *.greedo.net

Raspi — Raspberry Pi 5
└── k3s         (10.13.1.164)  — *.raspi.greedo.net
```

AdGuard (running on Altair) resolves `*.greedo.net → 10.13.1.167` and `*.raspi.greedo.net → 10.13.1.164`, with per-service overrides for Raspi services that don't use the `raspi.` subdomain.

## Repo layout

```
k8s/
  charts/
    homelab-app/       — shared Helm chart used by all services
  altair/
    apps/              — HelmRelease resources for each service
    infrastructure/    — configs, secrets, sources
  raspi/
    apps/              — HelmRelease resources for each service
    infrastructure/    — configs, secrets, sources
scripts/
  init-postgres-users.sh  — restore CNPG app users after cluster recreate
```

## Prerequisites

```bash
brew install helm kubectl sops fluxcd/tap/flux age
```

Your age private key must exist at `~/Library/Application Support/sops/age/keys.txt`.

## Bootstrap

See `scripts/README.md` for maintenance scripts (e.g. PostgreSQL user init).

### k3s on the Pi

```bash
curl -sfL https://get.k3s.io | sh -
```

Copy `/etc/rancher/k3s/k3s.yaml` to `~/.kube/config` and set the server IP to `10.13.1.164`.

### k3s on Altair (Proxmox LXC)

See the LXC setup notes — requires `/dev/kmsg` passthrough, `/proc/sys` remount, and `--disable=cloud-controller` in the k3s config.

### Install the SOPS age key

```bash
kubectl create namespace flux-system
kubectl create secret generic sops-age \
  --namespace=flux-system \
  --from-file=age.agekey="$HOME/Library/Application Support/sops/age/keys.txt"
```

Repeat for each cluster context.

### Bootstrap Flux

Requires a fine-grained GitHub PAT scoped to this repo with Contents and Administration read/write.

```bash
# Raspi
flux bootstrap github --owner=bvdwalt --repository=homelab --path=k8s/raspi --personal

# Altair
flux bootstrap github --owner=bvdwalt --repository=homelab --path=k8s/altair --personal --context=altair
```

Force an immediate sync with:

```bash
flux reconcile kustomization flux-system --with-source
```

## Secrets

Secrets are SOPS-encrypted with an age key. Edit a secret with:

```bash
sops k8s/<cluster>/infrastructure/secrets/<name>.sops.yaml
```

The `.sops.yaml` creation rule applies automatically to files under `k8s/*/infrastructure/secrets/`.
