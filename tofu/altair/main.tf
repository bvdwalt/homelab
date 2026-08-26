resource "proxmox_virtual_environment_container" "altair" {
  vm_id     = var.vmid
  node_name = var.proxmox_node

  unprivileged  = false # bootstrap doc uses --unprivileged 0
  start_on_boot = true

  features {
    nesting = true
    keyctl  = true
  }

  initialization {
    hostname = var.hostname

    ip_config {
      ipv4 {
        address = var.ip_address
        gateway = var.gateway
      }
    }

    dns {
      servers = var.dns_servers
      domain  = " " # matches live searchdomain (literal single space)
    }

    user_account {
      keys = var.ssh_public_keys
    }
  }

  console {
    enabled   = true
    tty_count = 2
    type      = "tty"
  }

  network_interface {
    name   = "eth0"
    bridge = "vmbr0"
  }

  disk {
    datastore_id = var.storage_pool
    size         = var.rootfs_size_gb
  }

  cpu {
    cores = var.cores
  }

  memory {
    dedicated = var.memory_mb
    swap      = var.swap_mb
  }

  operating_system {
    template_file_id = var.template_file
    type              = "debian"
  }

  started = true

  lifecycle {
    # template_file_id isn't recorded on already-existing containers, and
    # user_account is a creation-time-only setting — both force a
    # destroy-and-recreate if Tofu ever diffs them against a live import.
    ignore_changes = [
      operating_system,
      initialization[0].user_account,
    ]
  }
}

# The Proxmox API has no fields for apparmor/cgroup2/mount.entry/cap.drop/
# seccomp raw lines (see README) — append them to the conf file directly and
# restart, matching the manual bootstrap step plus everything added ad hoc
# since (GPU render node renumbered to card1, tun device for VPN, media bind
# mounts). Idempotent: only appends lines that aren't already present, and
# only restarts when something changed.
resource "null_resource" "raw_lxc_config" {
  depends_on = [proxmox_virtual_environment_container.altair]

  triggers = {
    vmid   = var.vmid
    script = local.raw_config_lines_joined
  }

  connection {
    type = "ssh"
    host = var.proxmox_ssh_host
    user = "root"
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOT
      set -e
      CONF=/etc/pve/lxc/${var.vmid}.conf
      CHANGED=0
      %{for line in local.raw_config_lines~}
      grep -qxF '${line}' "$CONF" || { echo '${line}' >> "$CONF"; CHANGED=1; }
      %{endfor~}
      if [ "$CHANGED" = "1" ]; then
        pct reboot ${var.vmid}
      fi
      EOT
    ]
  }
}

locals {
  raw_config_lines = [
    "lxc.apparmor.profile: unconfined",
    "lxc.cgroup2.devices.allow: c 226:1 rwm",
    "lxc.cgroup2.devices.allow: c 226:128 rwm",
    "lxc.mount.entry: /dev/dri/card1 dev/dri/card1 none bind,optional,create=file",
    "lxc.mount.entry: /dev/dri/renderD128 dev/dri/renderD128 none bind,optional,create=file",
    "lxc.mount.entry: /dev/kmsg dev/kmsg none bind,optional,create=file",
    "lxc.cgroup2.devices.allow: c 1:11 rwm",
    "lxc.cgroup2.devices.allow: c 10:200 rwm",
    "lxc.mount.entry: /dev/net/tun dev/net/tun none bind,optional,create=file",
    "lxc.mount.entry: /cheetah/music mnt/media/music none bind,create=dir 0 0",
    "lxc.mount.entry: /cheetah/shows mnt/media/shows none bind,create=dir 0 0",
    "lxc.mount.entry: /cheetah/movies mnt/media/movies none bind,create=dir 0 0",
    "lxc.mount.entry: /cheetah/downloads mnt/downloads none bind,create=dir 0 0",
    "lxc.mount.entry: /cheetah/k8s-nfs mnt/k8s-nfs none bind,create=dir 0 0",
    "lxc.seccomp.profile: ",
    "lxc.cap.drop: ",
    "lxc.cap.drop: mac_admin mac_override sys_time sys_rawio",
  ]
  raw_config_lines_joined = join("\n", local.raw_config_lines)
}
