FROM ubuntu:24.04

ARG HOST_UID=1000
ARG HOST_GID=1000

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8

# ---------- base packages ----------
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl gnupg git openssh-client \
      jq ripgrep build-essential pkg-config \
      less locales sudo tzdata \
    && rm -rf /var/lib/apt/lists/*

# ---------- arch detection ----------
# BuildKit provides TARGETARCH. Map it to the release-artifact naming
# conventions we need below.
ARG TARGETARCH
RUN test -n "${TARGETARCH}" || (echo "TARGETARCH unset — build with BuildKit" >&2 && exit 1)

# ---------- yq (standalone binary, not a language tool) ----------
RUN curl -fsSL -o /usr/local/bin/yq \
      "https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${TARGETARCH}" \
 && chmod +x /usr/local/bin/yq

# =====================================================================
# Languages. Each block below is self-contained — add a new language by
# copying one of these blocks. Keep them independent so they can be
# reordered or removed without touching anything else.
# =====================================================================

# ---------- Node.js 22.x ----------
# Kept around because many repos assume a working node toolchain at
# build time (install scripts, prebuilds, tree-sitter, etc.). Not used
# to install claude.
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
 && apt-get install -y --no-install-recommends nodejs \
 && rm -rf /var/lib/apt/lists/*

# ---------- Go ----------
ARG GO_VERSION=1.23.4
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" \
      -o /tmp/go.tgz \
 && tar -C /usr/local -xzf /tmp/go.tgz \
 && rm /tmp/go.tgz
ENV PATH="/usr/local/go/bin:${PATH}" \
    GOPATH="/home/agx/go"

# =====================================================================
# Tools that depend on a language runtime.
# =====================================================================

# ---------- GitHub CLI ----------
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends gh \
 && rm -rf /var/lib/apt/lists/*

# =====================================================================
# User setup. uid/gid are mapped to host so bind-mount writes appear as
# the invoking user on macOS/Linux — no root-owned files.
# =====================================================================
# ubuntu:24.04 ships with a pre-created `ubuntu` user at UID 1000. Delete
# it first so `useradd -u 1000` works on Linux hosts where the invoker's
# UID is 1000.
RUN (userdel -r ubuntu 2>/dev/null || true) \
 && (groupadd -g "${HOST_GID}" agx 2>/dev/null \
      || groupmod -n agx "$(getent group ${HOST_GID} | cut -d: -f1)") \
 && useradd -m -u "${HOST_UID}" -g "${HOST_GID}" -s /bin/bash agx \
 && echo 'agx ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/agx \
 && chmod 0440 /etc/sudoers.d/agx

# /home/agx is bind-mounted from the host at runtime (holds .ssh,
# .config/gh, .claude, etc.). That means nothing the image writes into
# ~ survives the mount — claude has to live in a system path instead.
ENV HOME=/home/agx \
    PATH="/usr/local/go/bin:/home/agx/go/bin:${PATH}"

# entrypoint.sh is NOT copied into the image — it's bind-mounted via
# docker-compose.yml at /agx/image. That way edits to the entrypoint
# apply on the next container start without rebuilding the image.

# ---------- Claude Code CLI (official installer, relocated) ----------
# Installer writes to $HOME/.local/share/claude/versions/<version> and
# symlinks $HOME/.local/bin/claude. We run it as the agx user, then
# move the binary + symlink into /usr/local as root so they survive the
# runtime bind-mount over /home/agx.
#
# CLAUDE_CACHE_BUST is a knob for pulling the newest upstream claude
# without nuking the earlier layers. Pass `--claude-latest` to
# `agx build` (or `--build-arg CLAUDE_CACHE_BUST=<anything-new>`)
# and this layer — plus the relocate layer below — rebuilds.
ARG CLAUDE_CACHE_BUST=0
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

USER agx
WORKDIR /workspace

ENTRYPOINT ["/agx/image/entrypoint.sh"]
