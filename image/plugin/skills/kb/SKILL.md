---
name: kb
description: Read and write the agx knowledge base at /kb. Invoke whenever
  task-relevant prior engineering context, architectural decisions, conventions,
  gotchas, or learnings might exist — before planning, during context capture,
  and when learning something worth preserving.
---

# Knowledge Base

The KB lives at `/kb`.

## Read pattern

1. Read `/kb/INDEX.md` first — it lists all entries with one-line summaries.
2. `rg <keyword>` across `/kb/` for relevant prior context.
3. Read the specific files that match.

## Write pattern

- Append to the most relevant existing file, or create a new one if nothing fits.
- Update `/kb/INDEX.md` with a one-line entry: `- [title](file.md) — one-line hook`
- Keep entries factual and dated (`YYYY-MM-DD`). No speculation. Mark uncertainty with `[?]`.

## Only write if ALL of these are true

- A future agent would be **surprised** not to know this
- It is **not** derivable from the code or git history
- It is a decision, constraint, or hard-won learning — not current state
- It is not already captured in `/kb/`

## Scope

Engineering context only: architecture decisions, repo conventions, gotchas,
runbooks, domain rules, learnings from past tasks.

Not: personal preferences (those live in auto-memory), credentials,
anything that belongs in git.
