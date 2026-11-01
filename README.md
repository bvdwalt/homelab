# homelab

![Validate](https://github.com/bvdwalt/homelab/actions/workflows/validate.yaml/badge.svg)
![TruffleHog](https://github.com/bvdwalt/homelab/actions/workflows/trufflehog.yaml/badge.svg)
![Checks](https://img.shields.io/github/checks-status/bvdwalt/homelab/main)

k3s cluster managed with Flux, running on an HP Elite Mini 800 G9.

## Key components

| Component | Role |
|-----------|------|
| [Flux](https://fluxcd.io) | GitOps — all cluster state is reconciled from this repo |
| [Cilium](https://cilium.io) | CNI with native routing |
| [Traefik](https://traefik.io) | Ingress controller, TLS termination |
| [cert-manager](https://cert-manager.io) | Wildcard TLS certificates via Let's Encrypt + Cloudflare DNS-01 |
| [CNPG](https://cloudnative-pg.io) | CloudNativePG — shared PostgreSQL cluster on Altair |
| [AdGuard Home](https://adguard.com/adguard-home) | Local DNS with per-service rewrite rules |
| [Renovate](https://docs.renovatebot.com) | Automated dependency updates for Helm charts and container images |
| [SOPS + age](https://github.com/getsops/sops) | Secrets encrypted at rest in the repo |

## Architecture

```
Altair — HP Elite Mini 800 G9
├── Proxmox VE  (10.0.0.166)
└── k3s LXC     (10.0.0.167)  — *.greedo.net
```

AdGuard (running on Altair) resolves `*.greedo.net → 10.0.0.167`.

## Repo layout

```
k8s/
  charts/
    homelab-app/       shared Helm chart used by all services
  common/              shared Helm sources and the helmrelease-defaults component
  altair/
    flux-system/       Flux install, plus one Kustomization per app (apps/)
    apps/              per-app directories: HelmRelease, secrets, policies
    infrastructure/    configs, releases, sources, shared secrets
scripts/
  init-postgres-users.sh  restore CNPG app users after cluster recreate
```

## Prerequisites

```bash
brew install helm kubectl sops fluxcd/tap/flux age
```

Your age private key must exist at `~/Library/Application Support/sops/age/keys.txt`.

## Bootstrap

See `scripts/README.md` for maintenance scripts (e.g. PostgreSQL user init).

### k3s on Altair (Proxmox LXC)

See the LXC setup notes — requires `/dev/kmsg` passthrough, `/proc/sys` remount, and `--disable=cloud-controller` in the k3s config.

### Install the SOPS age key

```bash
kubectl create namespace flux-system
kubectl create secret generic sops-age \
  --namespace=flux-system \
  --from-file=age.agekey="$HOME/Library/Application Support/sops/age/keys.txt"
```

### Bootstrap Flux

Requires a fine-grained GitHub PAT scoped to this repo with Contents and Administration read/write.

```bash
flux bootstrap github --owner=bvdwalt --repository=homelab --path=k8s/altair --personal --context=altair
```

Force an immediate sync with:

```bash
flux reconcile kustomization flux-system --with-source
```

## Secrets

Secrets are SOPS-encrypted with an age key. Edit a secret with:

```bash
sops k8s/altair/apps/<app>/<name>.sops.yaml
```

App secrets live next to their app. Secrets shared by infrastructure (OIDC tiers, Cloudflare, Crowdsec, Proxmox CA, Flux GitHub) live in `k8s/altair/infrastructure/secrets/`. The `.sops.yaml` creation rule applies to both.

## Adding an app

1. Add a Namespace with a pod-security label to `k8s/altair/apps/namespaces/namespaces.yaml`.
2. Create `k8s/altair/apps/<app>/` with a HelmRelease (`<app>.yaml`) and a `kustomization.yaml` that lists it and uses the `helmrelease-defaults` component. Copy `whoami` as a starting point.
3. Secrets: `sops k8s/altair/apps/<app>/<name>.sops.yaml`, list it in the app's `kustomization.yaml`, and reference it with `existingSecret`. Only secrets read by infrastructure belong in `infrastructure/secrets`.
4. Add `k8s/altair/flux-system/apps/<app>.yaml` with just the name and path, and list it in `k8s/altair/flux-system/apps/kustomization.yaml`. Interval, sops decryption and dependencies come from the shared patch there. Add an override in that patch if the app depends on something other than `namespaces`.
5. If you change the shared chart's templates or values, bump `version` in `Chart.yaml`. CI fails otherwise. Images must use a pinned tag or a digest.

## Removing an app

Namespaces and raw PVCs are labelled `kustomize.toolkit.fluxcd.io/prune: disabled`, so deleting their manifests leaves them in the cluster. Remove the app's files, then delete the namespace and PVCs with kubectl once the data is no longer needed.
