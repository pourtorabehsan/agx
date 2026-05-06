#!/usr/bin/env bash
set -euo pipefail

# Load any agent-visible settings (AGX_BRANCH_PREFIX, etc.)
if [ -f "$HOME/.agxrc" ]; then
  # shellcheck disable=SC1091
  . "$HOME/.agxrc"
fi

cd /workspace

mode="${AGX_MODE:-headless}"
agent="${AGX_AGENT:-codex}"
prompt_file="${AGX_PROMPT_FILE:-/workspace/.agx/prompt.txt}"

# Build optional --model flag if AGX_MODEL is set.
model_args=()
if [ -n "${AGX_MODEL:-}" ]; then
  model_args=(--model "$AGX_MODEL")
fi

codex_workspace_instructions() {
  if [ -f "$HOME/.codex/AGENTS.md" ] && [ ! -e /workspace/AGENTS.md ]; then
    cp "$HOME/.codex/AGENTS.md" /workspace/AGENTS.md
  fi
}

attach_prompt() {
  orient=""
  [ -s /workspace/.agx/prompt.txt ] && orient+="- /workspace/.agx/prompt.txt (original task)"$'\n'
  [ -s /workspace/CONTEXT.md ]       && orient+="- /workspace/CONTEXT.md (project context)"$'\n'
  [ -s /workspace/.agx/conduct.md ]  && orient+="- /workspace/.agx/conduct.md (session journal)"$'\n'
  if [ -n "$orient" ]; then
    printf 'You are resuming an agx workspace. Read these files to orient yourself, then wait for instructions:\n%s' "$orient"
  fi
}

case "$agent:$mode" in
  claude:headless)
    set +e
    claude -p "$(cat "$prompt_file")" "${model_args[@]}" --dangerously-skip-permissions
    rc=$?
    set -e
    echo "==> session exited: code=${rc} agent=${agent} mode=${mode}"
    exit "$rc"
    ;;
  codex:headless)
    codex_workspace_instructions
    set +e
    codex exec "${model_args[@]}" \
      --dangerously-bypass-approvals-and-sandbox \
      --skip-git-repo-check \
      --cd /workspace \
      "$(cat "$prompt_file")" \
      < /dev/null
    rc=$?
    set -e
    echo "==> session exited: code=${rc} agent=${agent} mode=${mode}"
    exit "$rc"
    ;;
  claude:interactive)
    if [ -s "$prompt_file" ]; then
      exec claude --dangerously-skip-permissions "${model_args[@]}" "$(cat "$prompt_file")"
    else
      exec claude --dangerously-skip-permissions "${model_args[@]}"
    fi
    ;;
  codex:interactive)
    codex_workspace_instructions
    if [ -s "$prompt_file" ]; then
      exec codex --dangerously-bypass-approvals-and-sandbox --cd /workspace "${model_args[@]}" "$(cat "$prompt_file")"
    else
      exec codex --dangerously-bypass-approvals-and-sandbox --cd /workspace "${model_args[@]}"
    fi
    ;;
  claude:attach)
    prompt="$(attach_prompt)"
    if [ -n "$prompt" ]; then
      exec claude --dangerously-skip-permissions "${model_args[@]}" "$prompt"
    else
      exec claude --dangerously-skip-permissions "${model_args[@]}"
    fi
    ;;
  codex:attach)
    codex_workspace_instructions
    prompt="$(attach_prompt)"
    if [ -n "$prompt" ]; then
      exec codex --dangerously-bypass-approvals-and-sandbox --cd /workspace "${model_args[@]}" "$prompt"
    else
      exec codex --dangerously-bypass-approvals-and-sandbox --cd /workspace "${model_args[@]}"
    fi
    ;;
  claude:resume)
    exec claude --dangerously-skip-permissions "${model_args[@]}"
    ;;
  codex:resume)
    codex_workspace_instructions
    exec codex --dangerously-bypass-approvals-and-sandbox --cd /workspace "${model_args[@]}"
    ;;
  *)
    echo "agx-entrypoint: unknown AGX_AGENT=$agent AGX_MODE=$mode" >&2
    exit 2
    ;;
esac
