---
name: kb
description: Read and write the agx knowledge base at /kb. Invoke whenever
  task-relevant prior engineering context, architectural decisions, conventions,
  gotchas, or learnings might exist — before planning, during context capture,
  and when learning something worth preserving. Also records and applies the
  user's engineering style and preferences so each session starts knowing them.
---

# Knowledge Base

The KB lives at `/kb`.

## Read pattern

1. Read `/kb/INDEX.md` first — it lists all entries with one-line summaries.
2. **Always read `/kb/user/profile.md`** before planning any non-trivial task — it captures the user's accumulated engineering style, preferences, and working habits. If it doesn't exist yet, skip silently.
3. `rg <keyword>` across `/kb/` for relevant prior context.
4. Read the specific files that match.

## Write pattern

- Append to the most relevant existing file, or create a new one if nothing fits.
- Update `/kb/INDEX.md` with a one-line entry: `- [title](file.md) — one-line hook`
- Keep entries factual and dated (`YYYY-MM-DD`). No speculation. Mark uncertainty with `[?]`.

## Only write if ALL of these are true

- A future agent would be **surprised** not to know this
- It is **not** derivable from the code or git history
- It is a decision, constraint, or hard-won learning — not current state
- It is not already captured in `/kb/`

## Categories

Suggested structure — use what fits, create new categories if none do:

- `decisions/` — architectural and technical choices with their reasoning
- `conventions/` — repo conventions, naming rules, patterns to follow
- `gotchas/` — non-obvious pitfalls, things that burned us before
- `runbooks/` — step-by-step procedures for recurring operations
- `repos/` — known repositories, their purpose, default branch, and relationships
- `architecture/` — system design, service boundaries, data flows, and infrastructure
- `user/` — user engineering profile written by the `reflect` skill

## Scope

Engineering context only: architecture decisions, repo conventions, gotchas,
runbooks, domain rules, learnings from past tasks, and the user profile
maintained by `reflect`.

Not: credentials, anything that belongs in git, ephemeral task state.
