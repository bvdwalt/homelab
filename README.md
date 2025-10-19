# Homelab Infrastructure (_cough_ Docker compose _cough_)

This repository contains Pulumi infrastructure code for managing homelab resources.

## Setup

1. Install Pulumi CLI:
    ```fish
    curl -fsSL https://get.pulumi.com | sh
    ```

3. Configure Pulumi:
    ```fish
    pulumi login
    pulumi stack init homelab
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
