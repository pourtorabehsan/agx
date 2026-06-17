FROM ubuntu:24.04

ARG HOST_UID=1000
ARG HOST_GID=1000

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8

# Base tools the agent will reach for: ssh, git, jq, rg, basic shell
# utilities, sudo for any apt-get the agent decides it needs.
# build-essential + pkg-config are baked in because npm/pip/etc. native
# extensions fail loudly without them and agents hit this constantly.
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl gnupg git openssh-client \
      jq ripgrep build-essential pkg-config \
      less locales sudo tzdata \
    && rm -rf /var/lib/apt/lists/*

# BuildKit-provided arch (amd64 / arm64). Required by the Go install below.
ARG TARGETARCH
RUN test -n "${TARGETARCH}" || (echo "TARGETARCH unset — build with BuildKit" >&2 && exit 1)

# Node.js 22 — Claude Code and Codex both expect a working node.
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
 && apt-get install -y --no-install-recommends nodejs \
 && rm -rf /var/lib/apt/lists/*

# Go toolchain. Pinned to match this repo's go.mod; bump together.
ARG GO_VERSION=1.25.0
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" \
      -o /tmp/go.tgz \
 && tar -C /usr/local -xzf /tmp/go.tgz \
 && rm /tmp/go.tgz
ENV GOPATH="/home/agx/go"

# GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends gh \
 && rm -rf /var/lib/apt/lists/*

# ubuntu:24.04 ships with a pre-created `ubuntu` user at UID 1000; remove it
# so `useradd -u 1000` works on Linux hosts where the invoker is UID 1000.
RUN (userdel -r ubuntu 2>/dev/null || true) \
 && (groupadd -g "${HOST_GID}" agx 2>/dev/null \
      || groupmod -n agx "$(getent group ${HOST_GID} | cut -d: -f1)") \
 && useradd -m -u "${HOST_UID}" -g "${HOST_GID}" -s /bin/bash agx \
 && echo 'agx ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/agx \
 && chmod 0440 /etc/sudoers.d/agx

ENV HOME=/home/agx \
    PATH="/usr/local/go/bin:/home/agx/go/bin:${PATH}"

# Install Claude Code as the agx user, then relocate to /usr/local so the
# binary survives the runtime bind-mount over /home/agx.
#
# CLAUDE_CACHE_BUST forces a fresh upstream pull without invalidating the
# earlier layers. Pass `--build-arg CLAUDE_CACHE_BUST=$(date +%s)` to
# rebuild this layer (and the relocate layer below).
ARG CLAUDE_CACHE_BUST=5
USER agx
WORKDIR /home/agx
RUN echo "claude cache bust: ${CLAUDE_CACHE_BUST}" \
 && curl -fsSL https://claude.ai/install.sh | bash
USER root
RUN set -eux; \
    mkdir -p /usr/local/share/claude; \
    cp -a /home/agx/.local/share/claude/versions /usr/local/share/claude/; \
    version=$(ls /usr/local/share/claude/versions | sort -V | tail -1); \
    ln -sf "/usr/local/share/claude/versions/${version}" /usr/local/bin/claude; \
    rm -rf /home/agx/.local /home/agx/.claude /home/agx/.cache /home/agx/.npm /home/agx/.claude.json; \
    /usr/local/bin/claude --version

# Install Codex CLI. It lives under the global npm prefix, outside the
# bind-mounted runtime home.
RUN npm install -g @openai/codex \
 && codex --version

# Scripts and agent commands/skills baked into the image
COPY entrypoint.sh /usr/local/bin/agx-entrypoint
COPY bootstrap.sh  /usr/local/bin/agx-bootstrap
COPY agent-exec.sh /usr/local/bin/agx-agent-exec
RUN chmod +x /usr/local/bin/agx-entrypoint /usr/local/bin/agx-bootstrap /usr/local/bin/agx-agent-exec
COPY plugin/ /usr/local/share/agx/plugin/
RUN chmod +x /usr/local/share/agx/plugin/post_bootstrap.sh

USER agx
WORKDIR /workspace
CMD ["/usr/local/bin/agx-entrypoint"]
