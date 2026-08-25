variable "tofu_encryption_passphrase" {
  description = "State/plan encryption key. Set via TF_VAR_tofu_encryption_passphrase, decrypted from encryption-passphrase.sops.yaml at runtime — never put this in a .tfvars file."
  type        = string
  sensitive   = true
}

variable "proxmox_endpoint" {
  description = "Proxmox API URL, e.g. https://10.0.0.166:8006"
  type        = string
  default     = "https://10.0.0.166:8006"
}

variable "proxmox_api_token" {
  description = "Proxmox API token, format 'user@realm!tokenid=uuid'"
  type        = string
  sensitive   = true
}

variable "proxmox_insecure" {
  description = "Skip TLS verification (self-signed Proxmox cert)"
  type        = bool
  default     = true
}

variable "proxmox_node" {
  description = "Proxmox node name — confirmed via `pvesh get /nodes` (single-node install kept the installer's default, not 'altair')."
  type        = string
  default     = "pve"
}

variable "proxmox_ssh_host" {
  description = "SSH address for the Proxmox host, used for the raw-lxc-config workaround"
  type        = string
  default     = "10.0.0.166"
}

variable "vmid" {
  description = "LXC container ID"
  type        = number
  default     = 100
}

variable "hostname" {
  type    = string
  default = "altair"
}

variable "template_file" {
  description = "Volume ID of the downloaded LXC template, e.g. local:vztmpl/debian-12-standard_12.7-1_amd64.tar.zst"
  type        = string
}

variable "cores" {
  type    = number
  default = 16
}

variable "memory_mb" {
  description = "Matches live config — bootstrap.md's 8192 was outgrown"
  type        = number
  default     = 12288
}

variable "swap_mb" {
  type    = number
  default = 512
}

variable "rootfs_size_gb" {
  description = "Matches live config — bootstrap.md's 32G was outgrown, then grown again after the external disk migration onto cheetah"
  type        = number
  default     = 96
}

variable "storage_pool" {
  description = "ZFS pool 'cheetah', not local-lvm — see project-hp-mini memory"
  type        = string
  default     = "cheetah"
}

variable "dns_servers" {
  description = "LXC-level resolv.conf. Matches AdGuard's node-LB IP, same as k3s's own resolv-conf."
  type        = list(string)
  default     = ["10.0.0.167", "1.1.1.1"]
}

variable "ip_address" {
  description = "Static IP/CIDR for the LXC's eth0"
  type        = string
  default     = "10.0.0.167/24"
}

variable "gateway" {
  description = "Deco mesh gateway"
  type        = string
  default     = "10.0.0.2"
}

variable "ssh_public_keys" {
  description = "SSH public key(s) to seed into the container's root account"
  type        = list(string)
}
