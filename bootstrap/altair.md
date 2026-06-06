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
lxc.cgroup2.devices.allow: c 1:11 rwm
lxc.mount.entry: /dev/dri/card0 dev/dri/card0 none bind,optional,create=file
lxc.mount.entry: /dev/dri/renderD128 dev/dri/renderD128 none bind,optional,create=file
lxc.mount.entry: /dev/kmsg dev/kmsg none bind,optional,create=file
```

Add each line individually with `echo '...' >> /etc/pve/lxc/100.conf` — pasting heredocs in the Proxmox web shell adds unwanted indentation.

Start the container:

```bash
pct start 100
pct enter 100
```

## 3. Install k3s

Inside the LXC, install curl first then k3s:

```bash
apt update && apt install -y curl
curl -sfL https://get.k3s.io | sh -
```

Proxmox mounts `/proc/sys` read-only inside the container, which prevents the kubelet from starting. Fix with a systemd override and a k3s config:

```bash
mkdir -p /etc/rancher/k3s
cat > /etc/rancher/k3s/config.yaml << EOF
protect-kernel-defaults: false
kubelet-arg:
  - "protect-kernel-defaults=false"
disable-cloud-controller: true
resolv-conf: /etc/rancher/k3s/resolv.conf
EOF

# Custom resolv.conf without greedo.net search domain.
# Without this, pods inherit the search domain and github.com.greedo.net
# (which hits TrueNAS via wildcard DNS) takes priority over github.com.
cat > /etc/rancher/k3s/resolv.conf << EOF
nameserver 10.13.1.165
EOF

mkdir -p /etc/systemd/system/k3s.service.d/
cat > /etc/systemd/system/k3s.service.d/proc-sys.conf << EOF
[Service]
ExecStartPre=/bin/mount -o remount,rw /proc/sys
EOF

systemctl daemon-reload && systemctl restart k3s
```

Verify the node is Ready:

```bash
export PATH=$PATH:/usr/local/bin
k3s kubectl get nodes
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
  --from-file=age.agekey=~/.config/sops/age/keys.txt
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

Bootstrap commits `gotk-components.yaml` and `gotk-sync.yaml` to `k8s/altair/flux-system/`. If bootstrap fails with "gotk-sync.yaml not found", the `kustomization.yaml` is referencing the file before it exists — temporarily remove `gotk-sync.yaml` from the resources list, push, rerun bootstrap, then restore it.

After bootstrap, the Flux controllers will be Pending due to an uninitialized cloud provider taint (because we disabled the cloud controller). Remove it:

```bash
kubectl taint node altair node.cloudprovider.kubernetes.io/uninitialized:NoSchedule-
```

The `gotk-sync.yaml` uses HTTPS (not SSH) to clone from GitHub, with a token-based secret. If bootstrap generates an SSH URL, update it manually:

```bash
# In k8s/altair/flux-system/gotk-sync.yaml, change:
#   url: ssh://git@github.com/bvdwalt/homelab
# to:
#   url: https://github.com/bvdwalt/homelab
kubectl delete secret flux-system -n flux-system
kubectl create secret generic flux-system \
  --namespace=flux-system \
  --from-literal=username=bvdwalt \
  --from-literal=password=<github-pat>
kubectl apply -f k8s/altair/flux-system/gotk-sync.yaml
```

## 7. Add DNS rewrite in AdGuard

Add a rewrite for `whoami-altair.greedo.net` → Altair's IP to verify the stack is working end-to-end.
