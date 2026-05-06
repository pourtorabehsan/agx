#!/usr/bin/env bash
set -euo pipefail

say()   { printf '==> %s\n' "$*"; }
pause() { read -r -p "$1" _; }
confirm() {
  local answer
  read -r -p "$1 [y/N] " answer
  case "${answer,,}" in
    y|yes) return 0 ;;
    *) return 1 ;;
  esac
}

mkdir -p "$HOME/.ssh" "$HOME/.claude" "$HOME/.codex" "$HOME/.config/gh"
chmod 700 "$HOME/.ssh"

# --- 1. SSH key ---
if [ ! -f "$HOME/.ssh/id_ed25519" ]; then
  ssh_email=$(git config --global user.email 2>/dev/null || true)
  if [ -z "$ssh_email" ]; then
    read -r -p "Your email address (for SSH key): " ssh_email
    if [ -z "$ssh_email" ]; then
      echo "agx-bootstrap: email is required for SSH key" >&2
      exit 1
    fi
  fi
  say "Generating SSH key at ~/.ssh/id_ed25519"
  ssh-keygen -t ed25519 -N '' -C "$ssh_email" -f "$HOME/.ssh/id_ed25519" >/dev/null
  cat >"$HOME/.ssh/config" <<'EOF'
Host github.com
  IdentityFile ~/.ssh/id_ed25519
  IdentitiesOnly yes
EOF
  chmod 600 "$HOME/.ssh/config"

  say "Add this public key to https://github.com/settings/ssh/new :"
  echo
  cat "$HOME/.ssh/id_ed25519.pub"
  echo
  pause "Press Enter once the key is added..."
else
  say "SSH key already present, skipping."
fi

# --- 2. gh auth ---
if [ ! -f "$HOME/.config/gh/hosts.yml" ]; then
  say "Logging into GitHub via gh"
  gh auth login -p ssh --git-protocol ssh -w
else
  say "gh already authenticated, skipping."
fi

# --- 3. agent logins ---
if [ ! -f "$HOME/.claude/.credentials.json" ]; then
  if confirm "Do you want to login to Claude?"; then
    say "Logging into Claude"
    # NOTE: verify the exact login invocation against the installed claude CLI.
    # `claude /login` is the current best-effort default.
    claude /login
  else
    say "Skipping Claude login."
  fi
else
  say "Claude already authenticated, skipping."
fi

if [ ! -f "$HOME/.codex/auth.json" ]; then
  if confirm "Do you want to login to Codex?"; then
    say "Logging into Codex"
    codex login --device-auth
  else
    say "Skipping Codex login."
  fi
else
  say "Codex already authenticated, skipping."
fi

# --- 4. git identity ---
if [ ! -f "$HOME/.gitconfig" ]; then
  read -r -p "Git user.name: "  GIT_NAME
  read -r -p "Git user.email: " GIT_EMAIL
  git config --global user.name  "$GIT_NAME"
  git config --global user.email "$GIT_EMAIL"
  git config --global init.defaultBranch main
  say "Wrote ~/.gitconfig"
else
  say ".gitconfig already present, skipping."
fi

# --- 5. Branch prefix ---
need_prefix=1
if [ -f "$HOME/.agxrc" ] && grep -q '^export AGX_BRANCH_PREFIX=' "$HOME/.agxrc"; then
  need_prefix=0
fi
if [ "$need_prefix" -eq 1 ]; then
  default_prefix=""
  if command -v gh >/dev/null 2>&1; then
    default_prefix=$(gh api user --jq .login 2>/dev/null || true)
  fi
  read -r -p "Branch prefix for agent commits [${default_prefix}]: " prefix
  prefix="${prefix:-$default_prefix}"
  if [ -z "$prefix" ]; then
    echo "agx-bootstrap: refusing to write empty AGX_BRANCH_PREFIX" >&2
    exit 1
  fi
  touch "$HOME/.agxrc"
  # Strip any existing line, then append.
  grep -v '^export AGX_BRANCH_PREFIX=' "$HOME/.agxrc" >"$HOME/.agxrc.tmp" || true
  mv "$HOME/.agxrc.tmp" "$HOME/.agxrc"
  printf 'export AGX_BRANCH_PREFIX=%q\n' "$prefix" >>"$HOME/.agxrc"
  say "Wrote AGX_BRANCH_PREFIX=$prefix to ~/.agxrc"
else
  say "AGX_BRANCH_PREFIX already set, skipping."
fi

# --- 6. Install agent skills and session defaults ---
if [ -d "$HOME/.agents/skills/conduct" ] || [ -L "$HOME/.claude/skills/conduct" ]; then
  say "Reinstalling agx skills into ~/.agents/skills and relinking Claude"
else
  say "Installing agx skills into ~/.agents/skills and linking Claude"
fi

mkdir -p "$HOME/.agents/skills" "$HOME/.claude/skills"
for skill_dir in /usr/local/share/agx/plugin/skills/*; do
  [ -d "$skill_dir" ] || continue
  skill_name=$(basename "$skill_dir")
  rm -rf "$HOME/.agents/skills/$skill_name"
  cp -R "$skill_dir" "$HOME/.agents/skills/$skill_name"

  rm -rf "$HOME/.claude/skills/$skill_name"
  ln -s "../../.agents/skills/$skill_name" "$HOME/.claude/skills/$skill_name"
done

cp /usr/local/share/agx/plugin/CLAUDE.md "$HOME/.claude/CLAUDE.md"
cp /usr/local/share/agx/plugin/AGENTS.md "$HOME/.codex/AGENTS.md"

# --- 7. Post-bootstrap hook ---
post=/usr/local/share/agx/plugin/post_bootstrap.sh
if [ -f "$post" ]; then
  say "Running post_bootstrap.sh"
  # shellcheck source=/dev/null
  source "$post"
fi

say "Bootstrap complete."
