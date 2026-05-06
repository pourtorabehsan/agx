#!/usr/bin/env bash
# Run after bootstrap completes. Sourced by agx-bootstrap every time it runs.
# Add any setup you want on every bootstrap: skill installs, config tweaks, etc.
#
# Example:
#   npx skills add https://github.com/vercel-labs/skills --skill find-skills

npx skills add https://github.com/vercel-labs/skills --skill find-skills --global --all --yes
npx skills add https://github.com/vercel-labs/agent-skills --skill vercel-react-best-practices --global --all --yes
npx skills add https://github.com/vercel-labs/agent-skills --skill web-design-guidelines --global --all --yes
npx skills add https://github.com/vercel-labs/agent-skills --skill vercel-composition-patterns --global --all --yes
npx skills add https://github.com/vercel-labs/agent-skills --skill vercel-react-view-transitions --global --all --yes
npx skills add https://github.com/vercel-labs/next-skills --skill next-best-practices --global --all --yes
npx skills add https://github.com/vercel-labs/next-skills --skill next-cache-components --global --all --yes
npx skills add https://github.com/shadcn/ui --skill shadcn --global --all --yes
npx skills add https://github.com/github/awesome-copilot --skill gh-cli --global --all --yes
npx skills add https://github.com/github/awesome-copilot --skill git-commit --global --all --yes