# Ansible Playbooks

Provisions the Thalos VM on TrueNAS SCALE and bootstraps the Kubernetes cluster.
All playbooks run on localhost and talk to TrueNAS via its REST API or to Talos via `talosctl`.

## Prerequisites

```bash
brew install talosctl talhelper helm kubectl sops
```

**Export required environment variables**

```bash
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8
export TRUENAS_API_KEY="your-key-here"
export CLOUDFLARE_API_TOKEN="your-token-here"   # DNS:Edit for greedo.net zone
```

Your age key must exist at `~/Library/Application Support/sops/age/keys.txt` for SOPS
decryption of `k8s/talos/talsecret.sops.yaml`.

## Playbooks

### Provision VM

Creates the Thalos zvol and VM on TrueNAS, uploads the Talos ISO, and starts the VM.
The VM boots into Talos maintenance mode and awaits configuration.

```bash
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-vm.yml
```

### Bootstrap cluster

Renders Talos machine config via talhelper, applies it to the node, bootstraps etcd,
fetches kubeconfig, then installs: cert-manager, democratic-csi, Traefik, whoami, and it-tools.

```bash
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-bootstrap.yml
```

Safe to re-run — all steps are idempotent.

### Tear down

Stops and deletes the VM and its zvol. Removes local `clusterconfig/` and `kubeconfig`.
Keeps `talsecret.sops.yaml` so the same cluster CA is reused on the next bootstrap.
Delete it manually for a completely fresh cluster identity.

```bash
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-teardown.yml
```
