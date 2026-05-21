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
2. **Always read `/kb/user/profile.md`** at the start of any non-trivial task — it captures the user's engineering style, preferences, and working habits accumulated across sessions.
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
- `user/` — the user's engineering style, preferences, and working habits

## User profile (`/kb/user/profile.md`)

This file is the persistent model of the user's engineering identity. Update it
whenever you observe a preference, correction, or pattern during the session.

**What to capture:**

- Corrections: when the user redirects your approach ("no, do it this way instead")
- Confirmations: when the user explicitly approves a non-obvious choice you made
- Stated preferences: opinions on tooling, architecture, naming, abstractions, review style
- Recurring patterns: things the user always/never does across multiple tasks
- Communication style: how they like explanations, what level of detail they want

**Structure of `/kb/user/profile.md`:**

```markdown
# User Engineering Profile

## Style & Approach
[How they like to work — concise vs. thorough, top-down vs. bottom-up, etc.]

## Code Preferences
[Language idioms, abstraction level, comment policy, naming conventions, etc.]

## What to Avoid
[Approaches they have rejected, anti-patterns they dislike]

## What Works Well
[Approaches they have confirmed or praised]

## Communication
[How they prefer explanations, feedback, and summaries]
```

**Write rules for profile entries:**

- Lead with the rule/preference, follow with **Why:** (their stated reason or inferred from context) and **Observed:** (the date and task where you saw it).
- Record from both corrections AND confirmations — only capturing corrections creates an overly cautious, tentative profile.
- If an existing entry conflicts with new evidence, update it rather than appending a contradiction.

## Scope

Engineering context and user working style: architecture decisions, repo conventions,
gotchas, runbooks, domain rules, learnings from past tasks, and the user's
accumulated engineering preferences and style.

Not: credentials, anything that belongs in git, or ephemeral task state.
