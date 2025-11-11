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

This project uses **Pulumi ESC (Environments, Secrets, and Configuration)** for centralized configuration management. All environment variables are stored in the `homelab-dev` Pulumi environment instead of local `.env` files or GitHub repository secrets.

### Managing Secrets

```fish
# View current environment
pulumi env open bvdwalt/homelab-dev

# Edit environment interactively
pulumi env edit bvdwalt/homelab-dev

# Set a specific value
pulumi env set bvdwalt/homelab-dev beszel.key "new-key-value" --secret
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
