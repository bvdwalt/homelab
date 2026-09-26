# netboot.xyz

Before Flux reconciles this app, create the ISO host directory on Altair
(the LXC), since it's mounted as a hostPath rather than a PVC:

```bash
mkdir -p /mnt/isos
```

Self-hosted ISOs placed there are served at `/assets` inside the container
per netboot.xyz's own volume convention.
