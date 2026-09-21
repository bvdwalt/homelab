# Claude Code

Day to day: `claude rc` (Remote Control) runs in a pod with a persistent
clone of this repo, so the phone app's Code tab drives it directly.
Break-glass: a `ttyd` terminal at `https://claude.greedo.net`, behind the
`infra-admin` Pocket ID tier — only needed to re-run `claude auth login`
when the Remote Control login expires (it can't be renewed remotely, and
that flow needs an interactive terminal). PVC at `/home/claude` holds the
repo, shell history, and login credentials. `git push` auth comes from a
GitHub PAT in the `claude-code` Secret. Read-only in-cluster `kubectl` via
`rbac.yaml` (`view` ClusterRole). Personal `CLAUDE.md`, `settings.json`, and
skills are pulled from `github.com/bvdwalt/dotfiles` (`home/dot_claude/`)
into `~/.claude` on every pod start — edit them there, not on the pod.

## One-time setup

1. **Login** — `claude rc` needs a claude.ai login before it'll start, so
   the pod retries in the background until you do it, either from the
   `ttyd` terminal (`claude auth login`) or:

   ```bash
   kubectl exec -it -n claude-code deploy/claude-code -- claude auth login
   ```

   Follow the printed URL. The next retry (within ~30s) picks up the
   session automatically — no restart needed. It'll show up in the phone
   app's Code tab as `homelab`. Repeat this whenever the login expires.

2. **GitHub PAT** — fine-grained token scoped to `bvdwalt/homelab`,
   Contents: read/write. Then, with `sops`/`age` set up locally:

   ```bash
   cat <<'EOF' > k8s/altair/infrastructure/secrets/claude-code.sops.yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: claude-code
     namespace: claude-code
   stringData:
     GITHUB_TOKEN: <paste PAT here>
   EOF
   sops -e -i k8s/altair/infrastructure/secrets/claude-code.sops.yaml
   ```

   Add it to `k8s/altair/infrastructure/secrets/kustomization.yaml`, commit, push.

3. **First image build** — `images/claude-code/` builds on push to `main`
   or via `workflow_dispatch`. Until it's built once, `claude-code.yaml`'s
   `image:` is a placeholder — update it with the resulting
   `ghcr.io/bvdwalt/homelab-claude-code:sha-<commit>@sha256:<digest>`.
