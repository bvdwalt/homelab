# netboot.xyz

The `/mnt/isos` hostPath is backed by a bind mount declared in
`tofu/altair/main.tf` (`/cheetah/isos` → `mnt/isos`), not created by hand.

Before applying, create the host-side directory once on the Proxmox host
(10.0.0.166) — `create=dir` in the mount entry only creates the LXC-side
mount point, not the ZFS-side source:

```bash
mkdir -p /cheetah/isos
```

Then run `tofu apply` from `tofu/altair` (see that directory's README).

Self-hosted ISOs placed there are served at `/assets` inside the container
per netboot.xyz's own volume convention.
