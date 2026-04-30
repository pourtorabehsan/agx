---
name: reflect
description: Post-session retrospective that distills durable learnings from the session into the knowledge base via the kb skill. Invoke at the end of every session that touched a codebase — even short ones. Don't skip it because the session felt routine; the most useful kb entries often come from tasks that seemed simple. If a session is wrapping up and reflect hasn't run, run it now.
---

You are an archivist, not an analyst. Your job is to decide what a future agent would need to know that they couldn't find in the code or git history — and write only that. The goal is a lean, high-signal knowledge base. One precise entry is worth more than five vague ones.

## Process

1. Read what the session produced:
   - `/workspace/CONTEXT.md` — what was known going in
   - `/workspace/PLAN.md` — what was planned (if it exists)
   - `git diff main` — what actually changed
   - Any errors, surprises, or pivots that came up during the session

2. For each candidate learning, ask: **"Would a future agent working on a similar task be surprised not to know this?"**
   - Yes → write it via the `kb` skill
   - No → skip it

   This question matters because the knowledge base is not a log — it's a resource. Entries that describe the obvious or the ephemeral just make the real entries harder to find.

3. Write findings via the `kb` skill. Prefer appending to an existing file over creating a new one — fragmented entries on the same topic are hard to cross-reference.

## What's worth writing

- Decisions and their reasoning ("we chose X over Y because Z constraint") — the code shows X but not why
- Non-obvious dependencies or ordering requirements that aren't visible in the code
- Gotchas discovered during implementation — especially things that wasted time
- Domain rules that live in someone's head, not in documentation
- Patterns likely to recur in future tasks

## What to skip

- Anything derivable from the code or git history — `git blame` gets there faster
- Current state ("the PR is merged", "the bug is fixed") — it will be wrong in a week
- Obvious things ("we used Go", "the test passed")
- Anything already in the knowledge base — check before writing
- Personal preferences that don't generalise

## Rules

- Be specific. Vague entries ("fixed a tricky bug") are noise that erodes trust in the kb.
- Date every new entry (`YYYY-MM-DD`) — entries without dates become uninterpretable as the codebase evolves.
- Mark uncertainty with `[?]` — a confidently wrong entry is worse than no entry.
- If nothing durable was learned this session, write nothing. An empty reflect is correct and honest.
- Do not invent learnings. Only write what actually happened.
