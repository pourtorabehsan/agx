# agx v1 — conductor

Date: 2026-04-24
Status: draft (brainstorm output)

## Purpose

v0 runs one Claude session in one container. v1 adds a **conductor**: a first Claude pass that triages the task, emits a structured workflow, and lets the runtime execute it — with optional parallelism, checkpoints that re-invoke the conductor, and a pause-to-interject mechanism.

The design priority is **portability**: the entire runtime must live inside the image so the image is usable on a remote host or CI box with no host dependency beyond Docker. That rules out spawning sibling containers per step (would require host Docker socket / DinD). Steps therefore run as child processes inside one container; kernel-level isolation is traded for mobility.

## Goals

- One image. `docker run agx:latest <prompt>` on any Docker host runs the full workflow — no `agx` binary, no host mounts beyond the workspace, no privileged flags.
- Conductor is itself a Claude session. It decides the workflow shape from the prompt and writes it down as JSON.
- Orchestrator is a Go binary baked into the image. It parses the workflow, executes steps with a bounded worker pool, re-invokes the conductor at declared checkpoints, and publishes status.
- Single-session v0 behavior (`AGX_MODE=run`) still works through the same binary — no bash entrypoint anymore.
- Programmatic interjection works the same locally and remotely: touch a file, orchestrator pauses at the next step boundary and asks the conductor to re-plan the tail.
- Observability from outside the container is a single `docker exec … cat /workspace/.agx/status.json`.

## Non-goals (deferred)

- Sibling / nested containers per step. Kernel isolation between steps is not offered.
- Preempting a running step mid-execution. Interjection takes effect at step boundaries.
- Durable workflow engine (Temporal/Airflow-style). No retries with exponential backoff, no persistent queue, no cross-container coordination.
- Live supervisor pattern (conductor consulted after every step). Only declared checkpoints.
- Live streaming of step output to the host in `conduct` mode. Per-step logs go to files; the orchestrator's own stdout is the merged top-level log.
- Session resume between steps (`claude --resume`). Every step is a fresh Claude session; state transfers via `/workspace` files.
- Cross-workflow caching. Re-running a workspace re-runs the conductor from scratch.
- Distributed execution (one workflow spanning multiple hosts).

## CLI surface

v0 commands unchanged. New:

```
agx conduct [flags] [prompt]
    -f, --file PATH       read prompt from file
        --name NAME       workspace name (default: slug of prompt)
    -i, --interactive     pause after conductor emits workflow.json; attach TTY so user can edit
        --workers N       parallel step cap (default: 2)
        --fail-fast       on step failure, abort workflow instead of asking conductor to re-plan
```

Prompt resolution and workspace name resolution follow v0 `agx run` rules.

`agx run` stays as the explicit single-session path. `agx conduct` is the workflow path. Both host-side commands are thin wrappers around `docker run` with different `AGX_MODE` env values; everything else lives in the image.

## Workflow JSON schema

Written by the conductor to `/workspace/.agx/workflow.json`:

```json
{
  "version": 1,
  "task": "short human-readable summary of the task",
  "steps": [
    {
      "id": "plan",
      "kind": "skill",
      "skill": "devbox:plan",
      "depends_on": [],
      "model": "opus",
      "tools": "default"
    },
    {
      "id": "review",
      "kind": "skill",
      "skill": "devbox:review",
      "depends_on": ["plan"],
      "model": "sonnet",
      "append_system_prompt": "You are reviewing an implementation plan. Be skeptical."
    },
    {
      "id": "impl_api",
      "kind": "prompt",
      "prompt": "Implement the API changes described in PLAN.md section 3.",
      "depends_on": ["review"],
      "model": "opus",
      "add_dir": ["/workspace/repos/api-bb"],
      "max_budget_usd": 10
    },
    {
      "id": "impl_ui",
      "kind": "prompt",
      "prompt": "Implement the UI changes described in PLAN.md section 4.",
      "depends_on": ["review"],
      "model": "opus",
      "add_dir": ["/workspace/repos/web"],
      "max_budget_usd": 10
    },
    {
      "id": "commit",
      "kind": "skill",
      "skill": "commit-commands:commit-push-pr",
      "depends_on": ["impl_api", "impl_ui"],
      "model": "haiku",
      "tools": ["Bash(git *)", "Bash(gh *)", "Read"]
    }
  ],
  "checkpoints": [
    {"after": "review", "mode": "replan_tail"}
  ]
}
```

Field rules (required):

- `steps[].id` — unique within the workflow; used as directory name under `.agx/steps/`.
- `steps[].kind` — `"skill"` or `"prompt"`.
- `steps[].skill` — skill slug; orchestrator expands to `claude -p "/<skill> …"` (exact form verified at impl time).
- `steps[].prompt` — raw prompt for `claude -p`.
- `steps[].depends_on` — must reference earlier step ids. DAG, cycle-free. Orchestrator topologically sorts and runs in waves bounded by `--workers`.
- `checkpoints[].after` — step id. When that step completes successfully, orchestrator halts the workflow, re-invokes the conductor with state context, and applies the reply.
- `checkpoints[].mode` — `"replan_tail"` is the only v1 mode. Conductor reply replaces all not-yet-started steps.

Field rules (optional per-step — map 1:1 to `claude` CLI flags):

- `model` — `"opus" | "sonnet" | "haiku"` (or a full model id). Default: `opus`. Orchestrator passes as `--model`.
- `effort` — `"low" | "medium" | "high" | "xhigh" | "max"`. Orchestrator passes as `--effort`.
- `tools` — `"default"` (all), `""` (none), or a list like `["Bash(git *)", "Edit", "Read"]`. Orchestrator passes as `--tools`. Narrowing this is the cheapest safety lever even under `--dangerously-skip-permissions`.
- `append_system_prompt` — step-specific role guidance appended to the default Claude Code system prompt. Orchestrator passes as `--append-system-prompt-file` (written to a file first).
- `system_prompt` — fully replace the default system prompt. Rare; implies the step doesn't want normal Claude Code behavior. Mutually exclusive with `append_system_prompt`. Implies `--bare` unless `bare: false` is explicit.
- `bare` — boolean. If true, passes `--bare` (skip hooks/LSP/plugin sync/auto-memory/CLAUDE.md auto-discovery). Default: false for steps, true for conductor invocations.
- `add_dir` — array of paths. Passed as `--add-dir`. Lets a step reach into sibling repos cloned elsewhere in `/workspace`.
- `max_budget_usd` — per-step cost cap. Passed as `--max-budget-usd`.
- `fallback_model` — passed as `--fallback-model`.
- `mcp_config` — path(s) to MCP server JSON. Passed as `--mcp-config`.
- `plugin_dir` — array of plugin dirs. Passed as `--plugin-dir` (repeated).

For a trivial task, the conductor emits a single-step workflow. No special-case code path.

## Conductor invocation

Both the initial conductor pass and checkpoint re-invocations use the same flag shape. The conductor is a pure function from (task + state) → structured JSON, so we lean on `--json-schema` for validation and `--bare` for determinism:

```
claude -p \
  --bare \
  --system-prompt-file /usr/local/share/agx/conductor.md \
  --output-format json \
  --json-schema "$CONDUCTOR_SCHEMA" \
  --tools "Read" \
  --dangerously-skip-permissions \
  --model opus \
  "$PROMPT_FILE_CONTENTS"
```

Why these choices:

- `--bare` — no CLAUDE.md auto-discovery, no auto-memory, no hooks, no plugin sync. The conductor must not inherit per-machine context; a given (task, state) input produces a reproducible workflow.
- `--system-prompt-file` (not `--append-system-prompt`) — the conductor's role replaces the default Claude Code system prompt entirely. It's not writing code, it's emitting JSON.
- `--json-schema` — Claude Code validates the output against a schema baked into the image at `/usr/local/share/agx/schemas/conductor.json`. Removes the brittle "parse this and hope" step. Invalid JSON → `claude -p` errors out → orchestrator aborts cleanly.
- `--output-format json` — single JSON object on stdout, not wrapped in chat framing.
- `--tools "Read"` — conductor can read state files; can't run Bash, can't Edit. Narrow blast radius.
- `--dangerously-skip-permissions` — still required because we're running in a sandbox; `--tools "Read"` does the actual restricting.
- `--model opus` — the triage-and-plan task is reasoning-heavy; don't downgrade.

### Initial pass

Input to the conductor is the user's original prompt plus (if present) the existing `/workspace/CONTEXT.md`. The schema returns a `workflow.json` object (same shape as the JSON Schema above). Orchestrator writes the result to `workflow.json` and `workflow-history/0.json`.

### Checkpoint re-invocation

At a checkpoint, the orchestrator:

1. Blocks until all steps up to and including the checkpoint step have finished.
2. Builds a prompt: original task + current `workflow.json` + summaries of completed step outputs (step id, skill/prompt, exit code, last ~50 lines of stdout, duration, cost).
3. Runs the conductor with a *checkpoint* JSON Schema whose top-level shape is:
   ```json
   {"decision": "continue"}
   {"decision": "replace_tail", "steps": [...], "checkpoints": [...]}
   {"decision": "abort", "reason": "..."}
   ```
4. On `continue`, proceeds with the existing tail. On `replace_tail`, rotates the prior `workflow.json` to `workflow-history/<n>.json`, installs the new one, and continues. On `abort`, stops cleanly with non-zero exit.

Both schemas live in the image (`/usr/local/share/agx/schemas/{conductor,checkpoint}.json`) and are versioned with it.

## Interjection

User (local or remote) touches `/workspace/.agx/interject` while the workflow is running. Before starting the next step, the orchestrator:

1. Consumes (deletes) the file.
2. Treats it as a synthetic checkpoint with `mode: replan_tail`.
3. Runs the conductor re-invocation flow above.

Remote equivalent: `docker exec <container> touch /workspace/.agx/interject`. No IPC, no port, no socket.

## `--interactive` mode

Orchestrator runs the initial conductor pass, writes `workflow.json`, then **pauses** and prints:

```
==> workflow ready: /workspace/.agx/workflow.json
==> press Enter to execute, or edit the file first
```

Host-side `agx conduct -i` adds `-it` to `docker run` so the user has a TTY to review/edit. For remote interactive use, `docker exec -it <container> bash` works the same way — the pause happens regardless.

`--interactive` only pauses at the initial workflow; checkpoint re-invocations proceed without pausing. (A `--interactive-checkpoints` flag could be added later if useful.)

## Status file

`/workspace/.agx/status.json`, rewritten atomically on every state transition:

```json
{
  "task": "implement vitess clone backup job",
  "mode": "conduct",
  "started_at": "2026-04-24T10:00:00Z",
  "workflow_version": 1,
  "worker_pool_size": 2,
  "interject_requested": false,
  "steps": [
    {"id": "plan",     "state": "completed", "started_at": "...", "ended_at": "...", "exit_code": 0},
    {"id": "review",   "state": "running",   "started_at": "...", "pid": 1234},
    {"id": "impl_api", "state": "pending"},
    {"id": "impl_ui",  "state": "pending"},
    {"id": "commit",   "state": "pending"}
  ]
}
```

States: `pending | running | completed | failed | skipped | aborted`.

Intended use: `watch -n1 'docker exec <c> cat /workspace/.agx/status.json | jq .'` from a remote shell, or `agx status <name>` on the host (thin wrapper over the same exec).

## Filesystem layout (additions to v0)

```
~/.agx/workspaces/<name>/.agx/
├── prompt.txt              # v0: user's original prompt
├── session.log             # v0: top-level orchestrator log (headless)
├── workflow.json           # current workflow (conductor output, mutated by replans)
├── status.json             # live orchestrator state
├── interject               # touch to pause at next step boundary (deleted after handled)
├── workflow-history/
│   ├── 0.json              # initial workflow
│   └── 1.json              # prior revision, after first replan
│   └── ...
└── steps/
    └── <step-id>/
        ├── prompt.txt          # resolved step prompt (skill expansion or raw)
        ├── system_prompt.txt   # only if workflow set append_system_prompt or system_prompt
        ├── stdout.log
        ├── stderr.log
        ├── events.jsonl        # only if step used stream-json; parsed events for cost/token accounting
        ├── meta.json           # exit_code, duration, timestamps, worker id, totals
        └── tmp/                # step-local TMPDIR
```

Image-side additions baked in at build time:

```
/usr/local/bin/agx                          # the in-container orchestrator (Go)
/usr/local/share/agx/
├── conductor.md                            # conductor system prompt template
├── checkpoint.md                           # checkpoint re-invocation system prompt template
└── schemas/
    ├── conductor.json                      # JSON Schema for workflow.json (initial pass)
    └── checkpoint.json                     # JSON Schema for checkpoint replies
```

## Step execution contract

Each step is a child process of the orchestrator. The orchestrator:

1. Creates `.agx/steps/<id>/` and `tmp/`.
2. Writes the resolved prompt to `prompt.txt`. For `kind: skill`, the resolved prompt is `"/<skill>"` (invocation form verified at impl time); for `kind: prompt`, it's the raw `prompt` field.
3. If the step sets `append_system_prompt` or `system_prompt`, writes it to `.agx/steps/<id>/system_prompt.txt`.
4. Builds a clean env: `HOME=/home/agx`, `PATH`, `TMPDIR=/workspace/.agx/steps/<id>/tmp`, `AGX_BRANCH_PREFIX` (from `~/.agxrc`), a per-step `CLAUDE_SESSION_NAME=<workspace>-<step-id>` (for the `--name` flag), explicit nothing else.
5. Builds the `claude` argv by mapping workflow fields to flags (see mapping below).
6. Starts the command in its own process group (`setpgid`).
7. Redirects stdout to `stdout.log`, stderr to `stderr.log`, and if `--output-format stream-json` is used, parses events into `events.jsonl` for cost/token accounting.
8. On step exit (or orchestrator cancellation), kills the entire process group (`SIGTERM` → 5s grace → `SIGKILL`). Guards against backgrounded dev servers leaking into the next step.
9. Writes `meta.json` with exit code, wall-clock duration, and (if stream-json was parsed) token/cost totals.

### Workflow field → claude flag mapping

| Workflow field            | claude flag                                        | Default if unset               |
|---------------------------|----------------------------------------------------|--------------------------------|
| `prompt` / resolved skill | positional prompt arg to `claude -p`               | —                              |
| `model`                   | `--model <v>`                                      | `opus`                         |
| `effort`                  | `--effort <v>`                                     | omitted                        |
| `tools`                   | `--tools <v>` (or `""` for none)                   | omitted (= default all)        |
| `append_system_prompt`    | `--append-system-prompt-file <path>`               | omitted                        |
| `system_prompt`           | `--system-prompt-file <path>` + `--bare` implied   | omitted                        |
| `bare`                    | `--bare`                                           | false for steps                |
| `add_dir`                 | `--add-dir <p1> <p2> ...`                          | omitted                        |
| `max_budget_usd`          | `--max-budget-usd <n>`                             | omitted                        |
| `fallback_model`          | `--fallback-model <v>`                             | omitted                        |
| `mcp_config`              | `--mcp-config <path...>`                           | omitted                        |
| `plugin_dir`              | `--plugin-dir <p>` (repeated)                      | omitted                        |
| *(always set)*            | `--print` / `-p`                                   | always                         |
| *(always set)*            | `--dangerously-skip-permissions`                   | always                         |
| *(always set)*            | `--name <CLAUDE_SESSION_NAME>`                     | always                         |
| *(always set)*            | `--output-format json`                             | always                         |
| *(always set)*            | `--no-session-persistence`                         | always (steps are one-shot)    |

`--no-session-persistence` keeps the per-user Claude session store from ballooning when workflows fan out widely. Resume semantics across steps are explicitly out of scope for v1; bring them back if/when a step needs multi-turn (`resume`/`session_id` fields can be added to the schema then).

Steps share `/workspace` and `/home/agx`. They do **not** share in-memory state — all coordination is file-based under `/workspace`.

## Worker pool

- Default cap: 2 concurrent steps. Configurable per run via `--workers N` (1–8 practical range).
- Scheduling: ready queue of steps whose `depends_on` are all `completed`. Pool pulls in submission order.
- Step failure: if `--fail-fast`, cancel all running steps (process-group kill) and exit non-zero. Otherwise, treat failure as a synthetic checkpoint — re-invoke the conductor with the failure context; the conductor decides whether to replan, abort, or retry (by emitting a new tail that includes a repeat of the failed step).
- Interjection is sampled between waves. A long-running step delays interjection until it exits — accepted v1 behavior.

## Image changes

- Replace bash `entrypoint.sh` with `/usr/local/bin/agx` (a second Go binary baked into the image, not the host CLI). It's the orchestrator.
- `Dockerfile` sets `CMD ["/usr/local/bin/agx"]` (bootstrap still invoked by overriding CMD with `agx-bootstrap`, as in v0).
- Dispatches on `AGX_MODE`:
  - `run` → v0 single-session behavior (still available, implemented in Go now).
  - `conduct` → full workflow flow.
  - `interactive` (v0) → folded into `run` with a TTY-detect.
- Reads `$HOME/.agxrc` at startup (parses `export KEY=VALUE` lines).
- Emits structured logs to `session.log` (conductor mode) or stdout (run mode).

The conductor system prompt lives as a template file baked into the image (e.g., `/usr/local/share/agx/conductor.md`) so it's versioned with the image.

## Host-side `agx conduct`

Go side:

1. Resolve prompt, workspace name, interactive flag, worker cap.
2. Write prompt to `<workspace>/.agx/prompt.txt`.
3. Exec docker:
   ```
   docker run --rm --init \
     [-it if --interactive] \
     -v ~/.agx/home:/home/agx \
     -v ~/.agx/workspaces/<name>:/workspace \
     -e AGX_MODE=conduct \
     -e AGX_WORKERS=<N> \
     -e AGX_FAIL_FAST=<0|1> \
     -e AGX_PROMPT_FILE=/workspace/.agx/prompt.txt \
     agx:latest
   ```
4. Tee stdout/stderr to `<workspace>/.agx/session.log` (headless path), no tee in interactive.
5. Exit code = container exit code.

Remote usage skips the host CLI entirely: `docker run --rm -v $(pwd)/ws:/workspace -e AGX_MODE=conduct -e AGX_PROMPT_FILE=/workspace/prompt.txt agx:latest` (host mounts may differ on CI; the point is the image works standalone).

## Internal structure (image-side Go)

New package tree under `image/orchestrator/` (built into `/usr/local/bin/agx` during `docker build`):

- `main.go` — dispatch on `AGX_MODE`.
- `run.go` — v0 single-session path.
- `conduct.go` — conductor-driven path: initial conductor pass, workflow loop, checkpoint handling, interjection.
- `workflow.go` — JSON schema, validator, topological sort.
- `pool.go` — bounded worker pool, dependency satisfaction, cancellation.
- `step.go` — step execution (process group, TMPDIR, log redirect, cleanup).
- `status.go` — atomic `status.json` writer.
- `agxrc.go` — parse `~/.agxrc` minimal `export` lines.

Shared with host-side Go? Not necessarily — the host CLI only builds docker args; duplication of workflow structs is tolerable. If it grows, promote to a shared `pkg/` package.

## Error handling

- Conductor emits invalid workflow.json (bad JSON, cycle, unknown skill, missing dep) → orchestrator errors with the validation failure, workflow is not started, exit non-zero.
- Step fails with `--fail-fast` → cancel siblings, mark pending steps `aborted`, exit non-zero.
- Step fails without `--fail-fast` → synthetic checkpoint, conductor re-invocation. If conductor fails too, hard-fail.
- Conductor re-invocation returns invalid JSON → treat as `abort` with reason `"conductor reply invalid"`. (No silent loop-retry in v1.)
- Interject file present but conductor unreachable → orchestrator continues, logs a warning.
- `status.json` write failure → log to `session.log`, keep running. Never fail the workflow over observability plumbing.

## Testing

Unit:

- Workflow JSON parser: valid cases, cycles, unknown deps, duplicate ids, unknown kinds.
- Topological sort stability.
- Worker pool: dependency satisfaction, cancellation propagation, `--fail-fast` vs not.
- Status writer: atomic rename, concurrent-read safety.
- `~/.agxrc` parser: handles quoted values, comments, unusual whitespace.

Integration (Docker required, skipped without):

- Stub conductor: a test image variant whose "conductor" binary just `cat`s a fixture `workflow.json` instead of calling Claude. Run end-to-end against a workflow that includes a fan-out, a checkpoint, and an interject. Assert: execution order respects deps, `--workers 2` actually parallelizes fan-out, interject triggers replan, `status.json` reflects final state.
- Per-step process-group kill: workflow with a step that backgrounds `sleep 1000`; assert no stray sleep processes survive step transition.

Real-Claude tests stay out of CI, same as v0.

## Open questions (implementation-time)

- Exact skill-invocation syntax for `claude -p`. `claude -p "/<skill> <args>"` is the current best guess; verify against the Claude Code CLI at impl time, and decide whether skill args come from the workflow or are hardcoded in the conductor's expansion.
- Conductor system prompt template — the biggest source of quality variance. First iteration goes in `/usr/local/share/agx/conductor.md`; expect several revisions post-ship. Same for the checkpoint template.
- `--json-schema` behavior under errors. What happens if the model can't produce schema-conforming JSON within the effort budget? Does `claude -p` exit non-zero with a clear error, or loop silently? Verify at impl time — if silent loop, we need a budget cap on the conductor itself.
- Whether to use `--output-format stream-json` + `--include-partial-messages` for steps so the orchestrator can do per-step token/cost accounting in `status.json` and enforce `max_budget_usd` ourselves, vs trusting `claude -p`'s own `--max-budget-usd` enforcement. Simpler to trust for v1; revisit if we want live cost dashboards.
- Should `agx status <name>` be a host command, or do we just document the `docker exec … cat status.json` pattern? Probably ship the host command; it's ~20 lines.
- Handling of Claude CLI upgrades: if `claude` flag surface changes, both `AGX_MODE=run` and step invocation break at once. Pin the `claude` version in the image; decide cadence for bumping. Defer the full policy to v1.1.
- Log volume: long-running workflows may produce large `stdout.log` files. v1 doesn't rotate. Flag for later if it bites.
- How the conductor learns the list of available skills. Options: (a) bake a skill catalog into the conductor prompt template, (b) have the orchestrator query `claude` for available skills and inject the list at runtime. (a) is simpler but stale-prone; (b) is cleaner. Lean (b) — verify the CLI supports skill listing.
- Use of `--exclude-dynamic-system-prompt-sections` for step invocations. Would improve prompt-cache reuse across parallel workers running similar steps, but only applies when the default system prompt is in use. Measure cache hit rate first; add if it matters.
- `--session-id` per step: worth setting deterministically (e.g. `<workspace-uuid>-<step-id>`) so a resumed workspace could in principle re-attach to prior sessions? Defer — steps are one-shot in v1 and `--no-session-persistence` is on, making this moot until we add multi-turn steps.
