FROM ubuntu:24.04

ARG HOST_UID=1000
ARG HOST_GID=1000

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8

# Base tools the agent will reach for: ssh, git, jq, rg, basic shell
# utilities, sudo for any apt-get the agent decides it needs.
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl gnupg git openssh-client \
      jq ripgrep less locales sudo tzdata \
    && rm -rf /var/lib/apt/lists/*

# Node.js 22 — several Claude Code subsystems expect a working node.
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
 && apt-get install -y --no-install-recommends nodejs \
 && rm -rf /var/lib/apt/lists/*

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

ENV HOME=/home/agx

# Install Claude Code as the agx user, then relocate to /usr/local so the
# binary survives the runtime bind-mount over /home/agx.
USER agx
WORKDIR /home/agx
RUN curl -fsSL https://claude.ai/install.sh | bash
USER root
RUN set -eux; \
    mkdir -p /usr/local/share/claude; \
    cp -a /home/agx/.local/share/claude/versions /usr/local/share/claude/; \
    version=$(ls /usr/local/share/claude/versions | sort -V | tail -1); \
    ln -sf "/usr/local/share/claude/versions/${version}" /usr/local/bin/claude; \
    rm -rf /home/agx/.local /home/agx/.claude /home/agx/.cache /home/agx/.npm /home/agx/.claude.json; \
    /usr/local/bin/claude --version

# Scripts baked into the image
COPY entrypoint.sh /usr/local/bin/agx-entrypoint
COPY bootstrap.sh  /usr/local/bin/agx-bootstrap
RUN chmod +x /usr/local/bin/agx-entrypoint /usr/local/bin/agx-bootstrap

USER agx
WORKDIR /workspace
CMD ["/usr/local/bin/agx-entrypoint"]
