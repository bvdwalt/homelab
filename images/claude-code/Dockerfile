FROM node:26-bookworm-slim@sha256:add5c17c14752c54d2f3d519ec46a3815684d41c33f198bcb09fb998f7dcbf68

ARG KUBECTL_VERSION=v1.33.5
ARG KUBECTL_SHA256=6a12d6c39e4a611a3687ee24d8c733961bb4bae1ae975f5204400c0a6930c6fc
ARG TTYD_VERSION=1.7.7
ARG TTYD_SHA256=8a217c968aba172e0dbf3f34447218dc015bc4d5e59bf51db2f2cd12b7be4f55
ARG GH_VERSION=2.103.3
ARG GH_SHA256=4c806978ffce1cec70853302a56d70ac6da4f19bae7adca4e1a8ab00633dd91d
ARG GO_VERSION=1.27.3
ARG GO_SHA256=65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40

RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
        curl \
        less \
        vim \
        procps \
        keepassxc \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
    && echo "${KUBECTL_SHA256}  /usr/local/bin/kubectl" | sha256sum -c - \
    && chmod +x /usr/local/bin/kubectl

RUN curl -fsSL -o /usr/local/bin/ttyd "https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}/ttyd.x86_64" \
    && echo "${TTYD_SHA256}  /usr/local/bin/ttyd" | sha256sum -c - \
    && chmod +x /usr/local/bin/ttyd

# gh reads GITHUB_TOKEN from the environment; no login step needed.
RUN curl -fsSL -o /tmp/gh.tar.gz "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz" \
    && echo "${GH_SHA256}  /tmp/gh.tar.gz" | sha256sum -c - \
    && tar -xzf /tmp/gh.tar.gz -C /tmp \
    && install -m 0755 "/tmp/gh_${GH_VERSION}_linux_amd64/bin/gh" /usr/local/bin/gh \
    && rm -rf /tmp/gh.tar.gz "/tmp/gh_${GH_VERSION}_linux_amd64"

# Go toolchain for developing Go tools on the pod. GOTOOLCHAIN=local stops go
# from trying to download a newer toolchain when a go.mod asks for one.
RUN curl -fsSL -o /tmp/go.tar.gz "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" \
    && echo "${GO_SHA256}  /tmp/go.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz
ENV PATH=/usr/local/go/bin:$PATH GOTOOLCHAIN=local

# renovate: datasource=npm depName=@anthropic-ai/claude-code
ARG CLAUDE_CODE_VERSION=2.1.325
RUN npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}

ENV HOME=/home/claude
RUN mkdir -p "$HOME" && chown node:node "$HOME"
USER node
WORKDIR /home/claude

COPY --chown=node:node entrypoint.sh /usr/local/bin/entrypoint.sh
USER root
RUN chmod +x /usr/local/bin/entrypoint.sh
USER node

EXPOSE 7681
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
