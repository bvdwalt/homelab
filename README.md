# homelab

Single-node Kubernetes cluster (Thalos) running on a TrueNAS SCALE VM, with Traefik for
ingress and democratic-csi for persistent storage backed by TrueNAS NFS.

## Architecture

```
TrueNAS SCALE (10.13.1.165)
└── Thalos VM (10.13.1.166) — Talos Linux, 4 vCPU, 8 GB
    ├── Traefik         — ingress, hostPort :80, *.k8s.home.lan
    ├── democratic-csi  — NFS StorageClass backed by Cheetah/k8s-nfs
    └── services/
        ├── it-tools
        └── whoami
```

AdGuard DNS on TrueNAS resolves `*.k8s.home.lan → 10.13.1.166`.

## Repo layout

```
ansible/      — provisions the Thalos VM and bootstraps the cluster
k8s/
  talos/      — talhelper config + SOPS-encrypted cluster secrets
  manifests/  — Helm values and Kubernetes manifests
```

## Prerequisites

```bash
brew install talosctl talhelper helm kubectl sops
export TRUENAS_API_KEY="your-key-here"
```

Your age key must exist at `~/Library/Application Support/sops/age/keys.txt`.

## Usage

```bash
# Provision VM + bootstrap cluster
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-vm.yml
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-bootstrap.yml

# Tear everything down
ansible-playbook -i ansible/inventory.yml ansible/playbooks/thalos-teardown.yml
```

See `ansible/README.md` for detailed playbook documentation.
