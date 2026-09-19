#!/bin/bash
set -euo pipefail

: "${GIT_USER_NAME:=Claude Code}"
: "${GIT_USER_EMAIL:=claude@localhost}"
: "${REPO_URL:=https://github.com/bvdwalt/homelab.git}"
REPO_DIR="${HOME}/homelab"

git config --global user.name "$GIT_USER_NAME"
git config --global user.email "$GIT_USER_EMAIL"
git config --global init.defaultBranch main
git config --global --add safe.directory "$REPO_DIR"

if [ -n "${GITHUB_TOKEN:-}" ]; then
    git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
fi

if [ ! -d "$REPO_DIR/.git" ]; then
    git clone "$REPO_URL" "$REPO_DIR"
fi

# In-cluster kubeconfig from this pod's own ServiceAccount token (read-only, see rbac.yaml)
SA_DIR=/var/run/secrets/kubernetes.io/serviceaccount
if [ -f "$SA_DIR/token" ]; then
    mkdir -p "$HOME/.kube"
    cat > "$HOME/.kube/config" <<EOF
apiVersion: v1
kind: Config
clusters:
  - name: in-cluster
    cluster:
      server: https://kubernetes.default.svc
      certificate-authority: $SA_DIR/ca.crt
contexts:
  - name: in-cluster
    context:
      cluster: in-cluster
      user: sa
current-context: in-cluster
users:
  - name: sa
    user:
      tokenFile: $SA_DIR/token
EOF
    export KUBECONFIG="$HOME/.kube/config"
fi

cd "$REPO_DIR"

# Retry claude rc in the background so it picks up login (or re-login) automatically.
(
    while true; do
        claude rc --name homelab || true
        sleep 30
    done
) &

# Break-glass terminal for `claude auth login`, e.g. when the rc login expires.
exec ttyd -p 7681 -W bash -l
