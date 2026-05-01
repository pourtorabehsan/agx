#!/usr/bin/env bash
set -euo pipefail

# Load any agent-visible settings (AGX_BRANCH_PREFIX, etc.)
if [ -f "$HOME/.agxrc" ]; then
  # shellcheck disable=SC1091
  . "$HOME/.agxrc"
fi

cd /workspace

mode="${AGX_MODE:-headless}"
prompt_file="${AGX_PROMPT_FILE:-/workspace/.agx/prompt.txt}"

case "$mode" in
  headless)
    claude -p "$(cat "$prompt_file")" --dangerously-skip-permissions
    rc=$?
    echo "==> session exited: code=${rc} mode=${mode}"
    exit "$rc"
    ;;
  interactive)
    if [ -s "$prompt_file" ]; then
      exec claude --dangerously-skip-permissions "$(cat "$prompt_file")"
    else
      exec claude --dangerously-skip-permissions
    fi
    ;;
  resume)
    exec claude --dangerously-skip-permissions
    ;;
  *)
    echo "agx-entrypoint: unknown AGX_MODE=$mode" >&2
    exit 2
    ;;
esac
