#!/usr/bin/env bash
# Run after bootstrap completes. Sourced by agx-bootstrap every time it runs.
# Add any setup you want on every bootstrap: skill installs, config tweaks, etc.
#
# Example:
#   npx skills add https://github.com/vercel-labs/skills --skill find-skills

npx skills add https://github.com/vercel-labs/skills --skill find-skills --global --agent claude-code --agent codex --yes
npx skills add https://github.com/vercel-labs/agent-skills --skill vercel-react-best-practices --skill web-design-guidelines --skill vercel-composition-patterns --skill vercel-react-view-transitions --global --agent claude-code --agent codex --yes
npx skills add https://github.com/vercel-labs/next-skills --skill next-best-practices --skill next-cache-components --global --agent claude-code --agent codex --yes
npx skills add https://github.com/shadcn/ui --skill shadcn --global --agent claude-code --agent codex --yes
npx skills add https://github.com/github/awesome-copilot --skill gh-cli --skill git-commit --global --agent claude-code --agent codex --yes
