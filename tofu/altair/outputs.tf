output "vmid" {
  value = proxmox_virtual_environment_container.altair.vm_id
}

output "ip_address" {
  value = var.ip_address
}
