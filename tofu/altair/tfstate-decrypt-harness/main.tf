# Minimal decrypt-only harness for the git diff textconv driver
# (../tfstate-diff.sh). Deliberately has no provider/resource config so
# `tofu init` here needs no network access — only `tofu state pull` needs
# to work. Keep the `encryption` block in sync with ../versions.tf if that
# ever changes.

terraform {
  encryption {
    key_provider "pbkdf2" "main" {
      passphrase = var.tofu_encryption_passphrase
    }

    method "aes_gcm" "main" {
      keys = key_provider.pbkdf2.main
    }

    state {
      method = method.aes_gcm.main
    }

    plan {
      method = method.aes_gcm.main
    }
  }
}

variable "tofu_encryption_passphrase" {
  type      = string
  sensitive = true
}
