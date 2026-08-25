# Pocket ID

## Tiered OIDC groups (per-app-group Traefik auth)

Every app protected by Pocket ID gets its own dedicated group + OIDC client
+ Traefik `Middleware` — a "tier" — rather than sharing one client/Middleware
across unrelated apps. This keeps blast radius scoped: a new tier, or
reassigning an app to a different tier, doesn't touch any other app's auth.
There is no shared/default `oidc-auth` Middleware anymore (decommissioned
once every app had a tier — see git history if you need the old pattern).

Existing tiers:

| Tier | Membership | Apps |
|------|------------|------|
| `downloads` | all household users | radarr, sonarr, lidarr, bazarr, prowlarr, qbittorrent, flaresolverr, profilarr, music-grabber, slskd |
| `household` | all household users | actual-budget, glance, jellystat, metering-dashboard, overdrive, seerr |
| `infra-admin` | admin only | adguard, backrest, it-tools, whoami, vaultwarden, traefik dashboard, hubble-ui |

`household` and `downloads` currently have identical membership but are
kept as separate Pocket ID groups/clients rather than merged, so that
granting someone downloads access doesn't implicitly grant household-app
access (and vice versa) — same scoping principle as the tiers themselves.

Reference implementation for a new tier: the `downloads` tier
(`k8s/altair/apps/downloads/ingressroutes.yaml`,
`k8s/altair/infrastructure/configs/downloads-oidc-auth.yaml`,
`k8s/altair/infrastructure/secrets/downloads-oidc-auth.sops.yaml`), or
`household`/`infra-admin` for apps using the shared `homelab-app` chart
(`middlewares:` override in the app's `HelmRelease` values, instead of a
separate `ingressroutes.yaml`).

### 1. Create a Pocket ID API key

Pocket ID has no bootstrap-token env var like Authentik — a key must be
created manually: **Settings → API Keys → New**, with client + group
management permissions. Use it as `X-API-KEY: <key>` (not `Authorization:
Bearer`).

### 2. Create the group and add members

```bash
curl -s -X POST -H "X-API-KEY: $KEY" -H "Content-Type: application/json" \
  https://pocketid.greedo.net/api/user-groups \
  -d '{"name": "<tier>", "friendlyName": "<Tier>"}'
# -> returns group id

curl -s -X PUT -H "X-API-KEY: $KEY" -H "Content-Type: application/json" \
  "https://pocketid.greedo.net/api/user-groups/<group-id>/users" \
  -d '{"userIds": ["<user-id>", ...]}'
```

### 3. Create the OIDC client, restrict it to the group

```bash
curl -s -X POST -H "X-API-KEY: $KEY" -H "Content-Type: application/json" \
  https://pocketid.greedo.net/api/oidc/clients \
  -d '{
    "name": "<Tier> Tier",
    "callbackURLs": ["https://*.greedo.net/*/callback"],
    "logoutCallbackURLs": ["https://*.greedo.net/*/callback"],
    "isPublic": false,
    "pkceEnabled": false
  }'
# -> returns client id
```

The wildcard callback URL already covers every `*.greedo.net` host, so no
per-app callback entries are needed — this mirrors the existing shared
`Traefik` client.

Group restriction is a **separate endpoint** — `isGroupRestricted` /
`allowedUserGroups` on the main `PUT /api/oidc/clients/{id}` do **not**
persist (confirmed by testing: the field silently reverts to empty). Use:

```bash
curl -s -X PUT -H "X-API-KEY: $KEY" -H "Content-Type: application/json" \
  "https://pocketid.greedo.net/api/oidc/clients/<client-id>/allowed-user-groups" \
  -d '{"userGroupIds": ["<group-id>"]}'
```

Then generate the client secret (shown once):

```bash
curl -s -X POST -H "X-API-KEY: $KEY" \
  "https://pocketid.greedo.net/api/oidc/clients/<client-id>/secret"
```

### 4. Generate a session secret

The `traefik-oidc-auth` plugin's `Secret` field must be **exactly 32
characters** — not base64 (`openssl rand -base64 32` produces 44 chars and
fails silently with `Invalid secret provided` in the Traefik logs). Use:

```bash
openssl rand -hex 16   # 32 hex chars
```

### 5. Repo changes

- `k8s/altair/infrastructure/secrets/<tier>-oidc-auth.sops.yaml` — new SOPS
  secret with keys `<TIER>_AUTH_CLIENT_ID`, `<TIER>_AUTH_CLIENT_SECRET`,
  `<TIER>_OIDC_SESSION_SECRET`. **Key names must be prefixed** — all
  `postBuild.substituteFrom` secrets referenced by `infrastructure-configs`
  share one variable namespace, so reusing the plain `AUTH_CLIENT_ID` etc.
  names from the shared `oidc-auth` secret will collide.
- Add it to `k8s/altair/infrastructure/secrets/kustomization.yaml`.
- Add it as a `substituteFrom` entry in
  `k8s/altair/flux-system/infrastructure-configs.yaml`.
- `k8s/altair/infrastructure/configs/<tier>-oidc-auth.yaml` — new
  `Middleware` in `kube-system`, same `traefik-oidc-auth` plugin shape as
  `oidc-auth.yaml`, referencing the `${<TIER>_...}` vars.
- Add it to `k8s/altair/infrastructure/configs/kustomization.yaml`.
- Point each app's IngressRoute at the new Middleware instead of the shared
  `oidc-auth` one. No extra callback route is needed — unlike Authentik's
  outpost pattern, this plugin handles the OIDC callback inline on the same
  host.

### 6. Deploy

```bash
flux reconcile kustomization infrastructure-secrets --with-source
flux reconcile kustomization infrastructure-configs
flux reconcile kustomization <app-kustomization>
```

Check `kubectl -n kube-system logs -l app.kubernetes.io/name=traefik` for
`[traefik-oidc-auth]` errors if a route 404s or 401s right after rollout —
that's almost always a bad `Secret` length or a variable substitution
collision, not an auth failure.

## Header bypass for non-browser clients (mobile apps, scripts)

For clients that can't do a browser OIDC redirect (e.g. the Actual Budget
iOS app), each tier `Middleware` sets:

```yaml
BypassAuthenticationRule: "Header(`X-Oidc-Bypass-Key`, `${OIDC_BYPASS_HEADER_SECRET}`)"
```

One shared secret for all four tiers:
`k8s/altair/infrastructure/secrets/oidc-bypass.sops.yaml`
(`OIDC_BYPASS_HEADER_SECRET`, decrypt with `sops -d`). Sending that header
value skips OIDC everywhere, so only use it for apps with their own auth
behind it (Actual Budget requires its own server password regardless).

## Gotcha: apps with their own strict CSRF/Referer checks (e.g. qBittorrent)

Some backends (qBittorrent's WebUI is the concrete case we hit) validate
the `Origin`/`Referer` header against their own idea of "self" and reject
anything else with **no config override**. Two things break this behind an
external-IdP redirect flow like Pocket ID or Authentik:

1. **Scheme mismatch** — Traefik terminates TLS and proxies plain HTTP to
   the backend, so the app sees `http://` while the browser's `Origin` is
   `https://`. Fix: enable HTTPS on the backend's own web server (reuse the
   `greedo-net-wildcard-tls` secret) and add a Traefik `ServersTransport`
   with `serverName: <host>`, then set `scheme: https` +
   `serversTransport: <name>` on the IngressRoute's service. See
   `qbittorrent-transport` in `downloads/ingressroutes.yaml`.
2. **Referer mismatch** — after an external-IdP redirect completes, the
   browser's `Referer` on the final request is legitimately the IdP's
   domain (standard redirect-chain behaviour, not a bug). An app with a
   strict Referer check will reject this. Fix: strip the header before it
   reaches the backend with a Traefik `headers` Middleware:
   ```yaml
   apiVersion: traefik.io/v1alpha1
   kind: Middleware
   spec:
     headers:
       customRequestHeaders:
         Referer: ""
   ```
   Chain it after the auth Middleware in the route's `middlewares:` list.

Check the app's own log file (not just `kubectl logs`, which may be too
low-verbosity) for the exact rejection reason — qBittorrent's
`/config/qBittorrent/logs/qbittorrent.log` logged the precise
`Origin header & Target origin mismatch` / `Referer header & Target origin
mismatch` messages that made both fixes obvious instead of guesswork.
