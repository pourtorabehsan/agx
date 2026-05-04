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

# Build optional --model flag if AGX_MODEL is set.
model_args=()
if [ -n "${AGX_MODEL:-}" ]; then
  model_args=(--model "$AGX_MODEL")
fi

case "$mode" in
  headless)
    claude -p "$(cat "$prompt_file")" "${model_args[@]}" --dangerously-skip-permissions
    rc=$?
    echo "==> session exited: code=${rc} mode=${mode}"
    exit "$rc"
    ;;
  interactive)
    if [ -s "$prompt_file" ]; then
      exec claude --dangerously-skip-permissions "${model_args[@]}" "$(cat "$prompt_file")"
    else
      exec claude --dangerously-skip-permissions "${model_args[@]}"
    fi
    ;;
  attach)
    orient=""
    [ -s /workspace/.agx/prompt.txt ] && orient+="- /workspace/.agx/prompt.txt (original task)"$'\n'
    [ -s /workspace/CONTEXT.md ]       && orient+="- /workspace/CONTEXT.md (project context)"$'\n'
    [ -s /workspace/.agx/conduct.md ]  && orient+="- /workspace/.agx/conduct.md (session journal)"$'\n'
    if [ -n "$orient" ]; then
      exec claude --dangerously-skip-permissions "${model_args[@]}" \
        "$(printf 'You are resuming an agx workspace. Read these files to orient yourself, then wait for instructions:\n%s' "$orient")"
    else
      exec claude --dangerously-skip-permissions "${model_args[@]}"
    fi
    ;;
  resume)
    exec claude --dangerously-skip-permissions "${model_args[@]}"
    ;;
  *)
    echo "agx-entrypoint: unknown AGX_MODE=$mode" >&2
    exit 2
    ;;
esac
