#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: agx-agent-exec [--model MODEL] PROMPT_FILE" >&2
  exit 2
}

model=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --model|-m)
      [ "$#" -ge 2 ] || usage
      model="$2"
      shift 2
      ;;
    --)
      shift
      break
      ;;
    -*)
      usage
      ;;
    *)
      break
      ;;
  esac
done

[ "$#" -eq 1 ] || usage
prompt_file="$1"
[ -f "$prompt_file" ] || { echo "agx-agent-exec: prompt file not found: $prompt_file" >&2; exit 2; }

model_args=()
if [ -n "$model" ]; then
  model_args=(--model "$model")
fi

agent="${AGX_AGENT:-codex}"
case "$agent" in
  claude)
    claude -p "$(cat "$prompt_file")" "${model_args[@]}" --dangerously-skip-permissions
    ;;
  codex)
    codex exec "${model_args[@]}" \
      --dangerously-bypass-approvals-and-sandbox \
      --skip-git-repo-check \
      --cd /workspace \
      "$(cat "$prompt_file")" \
      < /dev/null
    ;;
  *)
    echo "agx-agent-exec: unknown AGX_AGENT=$agent" >&2
    exit 2
    ;;
esac
