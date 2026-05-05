# agx

Run sandboxed [Claude Code](https://claude.com/claude-code) sessions in Docker.

`agx` is a small Go CLI that launches one container per session. Each session gets its own workspace directory; credentials, the knowledge base, and the Claude install are persisted on the host and bind-mounted in.

## Why

- **Isolation.** The agent runs with `--dangerously-skip-permissions` inside a container instead of on your host.
- **Reproducibility.** A single image (`agx:latest`) pins the toolchain — Node, Go, `gh`, `rg`, `claude` — so every session starts from the same place.
- **Persistence where it matters.** SSH keys, `gh` auth, Claude credentials, and a shared knowledge base survive across sessions; the rest of the container is thrown away.

## Install

Requires Docker and Go 1.25+.

```sh
make build      # builds the CLI to ./bin/agx, symlinks ~/bin/agx, and builds agx:latest
```

Make sure `~/bin` is on your `PATH`.

## First-time setup

```sh
agx bootstrap
```

Interactively walks through:

1. Generate an `ed25519` SSH key, print the public key for you to add to GitHub.
2. `gh auth login` (SSH protocol).
3. `claude /login`.
4. Set `git` user.name / user.email.
5. Pick a branch prefix for agent commits (stored in `~/.agx/home/.agxrc` as `AGX_BRANCH_PREFIX`).
6. Install the bundled Claude skills (capture, conduct, kb, reflect) and `CLAUDE.md` into `~/.agx/home/.claude/`.

Steps 1–5 are idempotent — re-running `bootstrap` skips anything already configured. Step 6 always reinstalls the skills, so re-running `bootstrap` after a `make build` picks up any skill changes.

## Usage

```sh
agx run "explain the auth flow in this repo"        # headless, prompt as positional arg
agx run -f prompt.md                                # prompt from a file
echo "summarise yesterday's PRs" | agx run          # prompt from stdin
agx run -i                                          # interactive TTY session
agx run -d "summarise yesterday's PRs"              # detached (background); logs to session.log
agx run -n my-task "..."                            # explicit workspace name
agx run -r owner/repo "fix the flaky test"          # target a specific repository (-r is repeatable)
agx run --model claude-opus-4-7 "..."              # model override
agx run --memory 8g --cpus 4 "..."                 # container resource limits
agx run --no-conduct "what files are here?"         # skip automatic conduct invocation

agx ps                                              # list all workspaces
agx logs <name>                                     # print session log
agx logs -f <name>                                  # tail session log
agx logs --conduct <name>                           # show conductor journal
agx logs --phase <phase> <name>                     # show output for a specific phase
agx phases <name>                                   # list available phase outputs
agx kill <name>                                     # stop a running container
agx wait <name>                                     # block until container exits (propagates exit code)
agx attach <name>                                   # open an interactive session in an existing workspace
agx resume <name>                                   # re-run workspace with its original prompt
agx resume -i <name>                                # resume interactively
agx resume -d <name>                                # resume detached
agx rm <name> [<name>...]                           # remove workspaces
agx prune                                           # remove all (with confirmation)
agx upgrade                                         # reinstall Claude skills from the current image
agx cd <name>                                       # print workspace path (pair with shell function)
```

To actually `cd` into a workspace, add this to your shell profile:

```sh
agxcd() { cd "$(agx cd "$@")"; }
```

Headless runs tee stdout/stderr into `<workspace>/.agx/session.log`. Interactive runs attach a TTY and skip the log file. Detached runs (`-d` / `--detach`) run the container in the background, printing the workspace name and log path to stderr on start, and are mutually exclusive with `--interactive`. Both `run` and `resume` support `-d`, `-i`, `--model`, `--memory`, and `--cpus`.

Every non-empty headless or interactive prompt automatically prepends `invoke conduct skill`, so the `conduct` skill is the default entry point for all task-oriented runs. Pass `--no-conduct` to skip this (useful for simple or exploratory prompts). Pass `-i` without a prompt to drop into a plain Claude session.

`--repo` / `-r` is repeatable; each use appends `Target repo: owner/repo` to the prompt. The `conduct` skill uses these to identify which repositories to clone during the `clone` phase.

`agx resume` re-runs a stopped workspace with its original prompt — it is not interactive by default. Use `-i` for a TTY session or `-d` to run it in the background. It refuses to start if the container is already running; use `agx kill` first.

If `--name` is omitted, the workspace name is a slug of the prompt (lowercased, non-alphanumeric → `-`, trimmed to 40 chars on a word boundary). If the prompt is empty, the name falls back to a `YYYY-MM-DD-HHMMSS` timestamp.

## What lives where

```
~/.agx/
├── home/         → bind-mounted as /home/agx in the container
│   ├── .ssh/         SSH keys
│   ├── .config/gh/   gh auth
│   ├── .claude/      Claude credentials, skills, CLAUDE.md
│   ├── .gitconfig
│   └── .agxrc        AGX_BRANCH_PREFIX, etc.
├── workspaces/   → one subdir per session, bind-mounted as /workspace
│   └── <name>/
│       └── .agx/
│           ├── meta.json         prompt, repos, start time, model
│           ├── prompt.txt
│           ├── session.log
│           ├── conduct.md        conductor journal
│           └── phases/           phase prompts and outputs
└── kb/           → bind-mounted as /kb (shared knowledge base)
    └── INDEX.md
```

The container runs as a non-root user whose UID/GID match the host invoker, so files written through the bind mounts come out owned by you.

## The plugin

The image bakes in four skills, installed into `~/.agx/home/.claude/` on every `bootstrap`:

- **`capture`** — gathers context from the kb, GitHub, Notion, and Slack into `/workspace/CONTEXT.md`. No code, no plan — just grounded context.
- **`reflect`** — post-session retrospective. Writes durable learnings to `/kb` via the `kb` skill.
- **`kb`** — read/write convention for the shared knowledge base at `/kb`. Reads `INDEX.md` first, greps for prior context, and appends new entries with a one-line index hook.
- **`conduct`** — autonomous engineering conductor. Given a task involving code changes, classifies complexity, builds a phase plan, and delegates implementation to focused sub-sessions via `claude -p`.

Source lives under `image/plugin/`; edits there require an image rebuild (`make build-image`) and a re-run of `agx bootstrap` to copy them into the host home.

## Layout

```
cmd/agx/         Go CLI (cobra) — bootstrap, run, resume, ps, logs, kill, attach, wait, phases, upgrade, rm, prune
image/           Docker image
  Dockerfile     ubuntu:24.04 + node + gh + claude
  entrypoint.sh  picks headless vs interactive vs resume based on AGX_MODE
  bootstrap.sh   the first-run wizard above
  plugin/        Claude skills baked in
    CLAUDE.md    agent rules (never push to main, use AGX_BRANCH_PREFIX)
    skills/
      capture/   SKILL.md
      conduct/   SKILL.md
      kb/        SKILL.md
      reflect/   SKILL.md
Makefile         build / build-cli / build-image / test / clean
```

## Tests

```sh
make test    # go test ./...
```
