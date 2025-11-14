# Homelab Infrastructure (_cough_ Docker compose _cough_)

This repository contains Pulumi infrastructure code for managing homelab resources.

## Setup

1. Install Pulumi CLI:
    ```fish
    curl -fsSL https://get.pulumi.com | sh
    ```

2. Configure Pulumi:
    ```fish
    pulumi login
    pulumi stack init homelab
    ```

3. Set up Pulumi ESC environment:
    ```fish
    # Copy the example environment template
    cp pulumi-env.yaml pulumi-env-local.yaml
    
    # Edit with your actual values
    # Replace all placeholder values in pulumi-env-local.yaml
    
    # Import your configuration into Pulumi ESC
    pulumi env edit bvdwalt/homelab-dev -f pulumi-env-local.yaml
    
    # Verify the environment was created
    pulumi env open bvdwalt/homelab-dev
    ```

## Configuration Management

This project uses **Pulumi ESC (Environments, Secrets, and Configuration)** for centralized configuration management. All environment variables are stored in Pulumi environments instead of local `.env` files or GitHub repository secrets.

### Miniflux configuration

The `miniflux` block in `pulumi-env.yaml` now holds the database name plus the administrator credentials and flags that the Miniflux image relies on. Set `adminPassword` using `fn::secret`, keep `adminUsername` secure, and adjust `runMigrations`, `createAdmin`, or `debug` only if you understand how the container bootstraps the database and admin user.

### Environments

- **`homelab-dev`**: Used for local development (hostname: `truenas.local`)
- **`homelab-prod`**: Used for CI/CD via GitHub Actions (hostname: `truenas` for Tailscale)

The stack configuration file `Pulumi.dev.yaml` references `homelab-dev` for local use, while GitHub Actions overrides to use `homelab-prod` via the `PULUMI_ENVIRONMENTS` variable.

### Managing Secrets

```fish
# View current environment
pulumi env open bvdwalt/homelab-dev

# Edit environment from file
pulumi env edit bvdwalt/homelab-dev -f pulumi-env-local.yaml

# Edit environment interactively
pulumi env edit bvdwalt/homelab-dev

# Set a specific value
pulumi env set bvdwalt/homelab-dev beszel.key "new-key-value" --secret

# View in shell export format
pulumi env open bvdwalt/homelab-dev --format shell
```

## Common Commands

### Preview changes
```fish
pulumi preview
```

### Apply changes
```fish
pulumi up
```

### Refresh state
Use when containers have been removed on the remote Docker host and state needs to be updated:
```fish
pulumi refresh
```

### View current stack
```fish
pulumi stack
```
