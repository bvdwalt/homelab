# Committed — non-secret only. proxmox_api_token and
# tofu_encryption_passphrase are sourced as TF_VAR_* env vars from
# SOPS-encrypted files instead (see README.md).

template_file   = "local:vztmpl/debian-12-standard_12.12-1_amd64.tar.zst"
ssh_public_keys = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB28Vw7NOlf1enGxU8Eyo5oAu79ysoutsC5O0z5l0WRj Proxmox"]
