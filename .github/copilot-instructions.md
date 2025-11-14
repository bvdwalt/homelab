# Homelab Copilot Instructions

This is a Go-based Pulumi project that manages Docker containers on a remote homelab server via SSH. It deploys self-hosted services (Linkwarden, Miniflux, Beszel) with Traefik reverse proxy routing.

## Architecture Overview

**Key Pattern**: Infrastructure-as-code for Docker containers deployed to a single remote host. Services are configured via Pulumi ESC (cloud-based config/secrets), not local env files.

- `main.go` - Entry point: loads config → images → creates remote Docker provider → instantiates services → provisions containers
- `internal/config/env.go` - Pulumi config wrapper. All secrets use `pulumi.StringOutput` (encrypted) vs plain strings
- `internal/docker/service.go` - `ContainerService` abstracts container creation (image pull + container with Traefik labels)
- `internal/docker/services.go` - `HomelabServices` pre-configured service definitions (Linkwarden, Miniflux, Beszel, etc.)
- `internal/docker/utils.go` - Traefik label generation for HTTP routing (`servicename.domain.com`)
- `image-versions.yaml` - Single source of truth for Docker image versions (managed by Renovate bot)

## Critical Workflows

### Configuration Management
**Always use Pulumi ESC**, never local `.env` files:
```fish
# View/edit environment
pulumi env open bvdwalt/homelab-dev
pulumi env edit bvdwalt/homelab-dev -f pulumi-env-local.yaml

# Set secrets
pulumi env set bvdwalt/homelab-dev beszel.key "value" --secret
```

Config values are read in `config.GetConfig()` using `cfg.Require()` (plain) or `cfg.RequireSecret()` (encrypted).

### Deployment Commands
```fish
pulumi preview  # Dry run
pulumi up       # Deploy changes
pulumi refresh  # Sync state when containers manually removed
```

### Adding New Services
1. Add image to `image-versions.yaml` (use tag + digest for security)
2. Create method in `HomelabServices` returning `ContainerConfig`
3. Instantiate in `main.go` services slice
4. Add config fields to `Config` struct if needed (with `RequireSecret` for sensitive data)

Example pattern from `services.go`:
```go
func (h *HomelabServices) ServiceName(dbPassword pulumi.StringInput) ContainerConfig {
    return ContainerConfig{
        Name:         "service-name",
        ServiceName:  "service-name",  // Used in Traefik routing
        ImageName:    h.Images.Images["service-name"],
        InternalPort: 8080,
        DomainName:   h.DomainName,
        Networks:     []string{"network-name"},
        Volumes: []VolumeMount{{
            HostPath:      filepath.Join(h.SSDPath, "docker-volumes/service-name"),
            ContainerPath: "/data",
        }},
        Environment: map[string]pulumi.StringInput{
            "KEY": pulumi.String("value"),
            "SECRET": dbPassword,  // Use pulumi.StringInput for secrets
        },
    }
}
```

## Project-Specific Conventions

### Network & Routing
- **All web services connect to `proxy` network** and get Traefik labels auto-generated
- Services become accessible at `<servicename>.<domain>`
- Use `ExtraLabels` for Traefik middlewares (e.g., `"traefik.http.routers.<servicename>.middlewares": "oidc-auth"`)
- Agents/system services may use `NetworkMode: "host"` instead (see BeszelAgent)

### Volume Path Patterns
Three storage tiers referenced via config:
- `SSDPath` - Fast storage for app data (`docker-volumes/<service>`)
- `HDDPath` - Bulk storage
- `ExternalPath` - External drives

Always use `filepath.Join(h.SSDPath, "docker-volumes/<service>")` for consistency.

### Secret Handling
- Config struct: Use `pulumi.StringOutput` for secrets (not `string`)
- Service methods: Accept `pulumi.StringInput` for secret params
- Environment vars: Use `pulumi.Sprintf()` to interpolate secrets safely
- Example: `pulumi.Sprintf("postgresql://%s:%s@%s", user, password, host)`

### Image Management
- **Never hardcode image tags in code** - use `h.Images.Images["key"]` referencing `image-versions.yaml`
- Renovate automatically updates images with digest pinning (`@sha256:...`)
- Format: `"service-name": "image:tag@sha256:digest"`
- ghcr.io preferred over Docker Hub where possible

## Testing & Development

**No test suite currently exists**. To validate changes:
1. Run `pulumi preview` to see planned changes
2. Use `pulumi up` on dev stack first
3. Check container logs on remote host: `ssh user@host docker logs <container-name>`

## Common Gotchas

- **Don't call `pulumi.String()` on secrets** - use them directly as `pulumi.StringInput`
- **Provider context required**: All Docker resources must include `pulumi.Provider(cs.provider)`
- **Restart policy default**: Containers use `"unless-stopped"` unless overridden
- **Network defaults**: Services default to `proxy` network if Networks/NetworkMode not specified
