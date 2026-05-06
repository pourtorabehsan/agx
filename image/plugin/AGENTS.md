# Rules

- Treat Docker as the isolation boundary. It is acceptable to use the agent CLI's dangerous/bypass mode inside agx, but do not assume host-level access.
- Preserve user work. Never discard, reset, overwrite, or revert changes you did not make unless explicitly instructed.
- Do not push directly to the default branch. Use the `push` skill for final commit, branch, push, and PR creation.
- Prefer the bundled skills for their domains: `conduct` for engineering tasks, `capture` for context gathering, `kb` for durable knowledge, `reflect` for post-session learnings, and `push` for PR creation.
- Keep durable project knowledge in `/kb`; keep session-local artifacts under `/workspace/.agx/`.
- When credentials, repository access, or external permissions are missing, record the blocker clearly instead of working around it unsafely.
