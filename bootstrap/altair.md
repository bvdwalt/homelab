# Altair Bootstrap

HP Elite Mini 800 G9 — Proxmox VE + k3s LXC.

## 1. Proxmox VE

Install via USB ISO from https://www.proxmox.com/en/downloads. Standard install, no special options needed.

## 2. Create the k3s LXC

Download a template (Debian 12 recommended) via the Proxmox UI, then create the container:

```bash
pct create 100 local:vztmpl/debian-12-standard_*.tar.zst \
  --hostname altair \
  --cores 4 \
  --memory 8192 \
  --rootfs local-lvm:32 \
  --net0 name=eth0,bridge=vmbr0,ip=dhcp \
  --unprivileged 0
```

Then edit `/etc/pve/lxc/100.conf` to add the required flags and `/dev/dri` passthrough for Intel QSV:

```
features: keyctl=1,nesting=1
lxc.apparmor.profile: unconfined
lxc.cgroup2.devices.allow: c 226:0 rwm
lxc.cgroup2.devices.allow: c 226:128 rwm
lxc.mount.entry: /dev/dri/card0 dev/dri/card0 none bind,optional,create=file
lxc.mount.entry: /dev/dri/renderD128 dev/dri/renderD128 none bind,optional,create=file
```

Start the container:

```bash
pct start 100
pct enter 100
```

## 3. Install k3s

Inside the LXC:

```bash
curl -sfL https://get.k3s.io | sh -
```

## 4. Configure kubectl on your laptop

Copy the kubeconfig from the LXC, replacing `127.0.0.1` with Altair's IP, and add it as the `altair` context:

```bash
# On the LXC
cat /etc/rancher/k3s/k3s.yaml
```

```bash
# On your laptop — merge into ~/.kube/config or save and set KUBECONFIG
# Change server: https://127.0.0.1:6443 → https://<altair-ip>:6443
# Change name/context to: altair
kubectx altair  # verify
```

## 5. Create the SOPS age secret

Flux needs the age key before it can decrypt secrets. The existing raspi age key is reused.

```bash
kubectl create namespace flux-system
kubectl create secret generic sops-age \
  --namespace=flux-system \
  --from-file=age.agekey=/path/to/age.key
```

The age key is stored in 1Password.

## 6. Bootstrap Flux

```bash
flux bootstrap github \
  --owner=bvdwalt \
  --repository=homelab \
  --branch=main \
  --path=k8s/altair \
  --personal
```

This generates `k8s/altair/flux-system/gotk-components.yaml` and `gotk-sync.yaml` and commits them to the repo.

## 7. Add DNS rewrite in AdGuard

Add a rewrite for `whoami-altair.greedo.net` → Altair's IP to verify the stack is working end-to-end.
