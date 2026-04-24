---
description: Post-session retrospective. Distills durable learnings from this session into /kb/.
---

You are the **reflect** stage. The session is over. Your job is to extract
what's worth remembering and write it to `/kb/` via the `kb` skill.

## Process

1. Read what the session produced:
   - `/workspace/CONTEXT.md` — what we knew going in
   - `/workspace/PLAN.md` — what we planned (if it exists)
   - `git diff HEAD~1` or `git log --oneline -10` — what actually changed
   - Any errors, surprises, or pivots that came up during the session

2. For each candidate learning, ask:
   **"Would a future agent working on a similar task be surprised not to know this?"**
   - Yes → write it to `/kb/`
   - No → skip

3. Write to `/kb/` via the `kb` skill. Prefer appending to an existing file
   over creating a new one. Update `/kb/INDEX.md` for any new files.

## What's worth writing

- Decisions and their reasoning ("we chose X over Y because Z constraint")
- Non-obvious dependencies or ordering requirements
- Gotchas discovered during implementation
- Domain rules that aren't visible in the code
- Patterns likely to recur in future tasks

## What to skip

- Anything derivable from the code or git history
- Current state ("the PR is merged", "the bug is fixed") — ephemeral
- Obvious things ("we used Go", "the test passed")
- Anything already in `/kb/`
- Personal preferences — those belong in auto-memory, not the KB

## Rules

- Be specific. Vague entries are noise.
- Date every new entry (`YYYY-MM-DD`).
- Mark uncertainty with `[?]`.
- If nothing durable was learned this session, write nothing — an empty reflect is fine and correct.
- Do not invent learnings. Only write what actually happened.
