---
description: Gather context for a task from kb, GitHub, Notion, and Slack. Writes /workspace/CONTEXT.md.
argument-hint: "<task description>"
---

You are the **context-capture** stage. The task is:

```
$ARGUMENTS
```

Your single deliverable is `/workspace/CONTEXT.md`. Do not write code, clone repos, or write a plan.

## Process

Pull from four sources in order. Every source must either produce findings or
be recorded as "consulted: no hits" in the footer — never skipped silently.

### 1. Knowledge base (`kb` skill)

- Invoke the `kb` skill.
- `rg` across `/kb/` for prior notes related to this task.
- Include relevant snippets in CONTEXT.md with file references.

### 2. GitHub (via `gh` CLI)

- Identify candidate repos from the task description and kb references.
- For each repo, determine the base branch via
  `gh api repos/<owner>/<name> --jq .defaultBranchRef.name` —
  do not hardcode `main`.
- Pull related issues/PRs with `gh issue list` / `gh pr list` / `gh search issues`.

### 3. Notion (via MCP)

- Search for task keywords.
- Fetch promising pages; prefer recent docs (check `last_edited_time`).
- Skip anything clearly superseded by a newer doc.

### 4. Slack (via MCP)

- Use `slack_search_public_and_private` with task-specific terms
  (feature names, error strings, ticket IDs) — not generic keywords.
- Read at most 3 threads via `slack_read_thread`; include the permalink for each.
- Skip threads older than 90 days unless the task explicitly references them.
- Prefer threads with a clear resolution over active debate.

## Output format

Write `/workspace/CONTEXT.md` with YAML frontmatter:

```markdown
---
task: "<verbatim task string>"
repos:
  - url: git@github.com:owner/name.git
    base: main
---

# Context: <short title>

## What the task is about

<2–5 sentences. Ground truth only. No speculation.>

## Prior knowledge (from kb)

- <bullet with /kb path + one-line summary>

## Related issues / PRs

- <bullet with gh url + one-line summary>

## Notion docs

- <bullet with title + url + one-line summary>

## Slack threads

- <bullet with permalink + channel + one-line summary>

## Open questions

- <each uncertainty as a bullet, tagged [?]>

## Sources consulted

- kb: <hits | "no hits">
- GitHub: <N repos · M issues/PRs | "no hits">
- Notion: <N docs | "no hits" | "MCP unreachable">
- Slack: <N threads | "no hits" | "MCP unreachable">
```

## Rules

- Stay factual. No speculation. Flag uncertainty with `[?]`.
- Do not write code, a plan, or clone repos.
- Repo URLs must use SSH form (`git@github.com:…`).
- Never silently skip a source — record it under "Sources consulted".
- When done, verify the file has valid YAML frontmatter.
