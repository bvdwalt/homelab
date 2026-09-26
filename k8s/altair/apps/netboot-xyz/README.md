# netboot.xyz

Backed by a bind mount in `tofu/altair/main.tf` (`/cheetah/isos` →
`mnt/isos`); `tofu apply` from `tofu/altair` creates the host-side directory
too, no manual step needed. ISOs placed there are served at `/assets` inside
the container, per netboot.xyz's own volume convention.
