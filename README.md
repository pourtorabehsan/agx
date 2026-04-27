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
6. Install the bundled Claude commands (`/agx:capture`, `/agx:reflect`) and the `kb` skill into `~/.agx/home/.claude/`.

Each step is idempotent — re-running `bootstrap` skips anything already configured.

## Usage

```sh
agx run "explain the auth flow in this repo"        # headless, prompt as positional arg
agx run -f prompt.md                                # prompt from a file
echo "summarise yesterday's PRs" | agx run          # prompt from stdin
agx run -i                                          # interactive TTY session
agx run -n my-task "..."                            # explicit workspace name

agx ls                                              # list workspaces
agx rm <name> [<name>...]                           # remove workspaces
agx prune                                           # remove all (with confirmation)
```

Headless runs tee stdout/stderr into `<workspace>/.agx/session.log`. Interactive runs attach a TTY and skip the log file.

If `--name` is omitted, the workspace name is a slug of the prompt (lowercased, non-alphanumeric → `-`, trimmed to 40 chars on a word boundary). If the prompt is empty, the name falls back to a `YYYY-MM-DD-HHMMSS` timestamp.

## What lives where

```
~/.agx/
├── home/         → bind-mounted as /home/agx in the container
│   ├── .ssh/         SSH keys
│   ├── .config/gh/   gh auth
│   ├── .claude/      Claude credentials, commands, skills
│   ├── .gitconfig
│   └── .agxrc        AGX_BRANCH_PREFIX, etc.
├── workspaces/   → one subdir per session, bind-mounted as /workspace
│   └── <name>/
│       └── .agx/
│           ├── prompt.txt
│           └── session.log
└── kb/           → bind-mounted as /kb (shared knowledge base)
    └── INDEX.md
```

The container runs as a non-root user whose UID/GID match the host invoker, so files written through the bind mounts come out owned by you.

## The plugin

The image bakes in two slash commands and a skill, installed into `~/.agx/home/.claude/` on first `bootstrap`:

- **`/agx:capture <task>`** — gathers context from the kb, GitHub, Notion, and Slack into `/workspace/CONTEXT.md`. No code, no plan — just grounded context.
- **`/agx:reflect`** — post-session retrospective. Writes durable learnings to `/kb` via the `kb` skill.
- **`kb` skill** — read/write convention for the shared knowledge base at `/kb`. Reads `INDEX.md` first, greps for prior context, and appends new entries with a one-line index hook.

Source lives under `image/plugin/`; edits there require an image rebuild (`make build-image`) and a re-run of `agx bootstrap` to copy them into the host home.

## Layout

```
cmd/agx/         Go CLI (cobra) — bootstrap, run, ls, rm, prune
image/           Docker image
  Dockerfile     ubuntu:24.04 + node + gh + claude
  entrypoint.sh  picks headless vs interactive based on AGX_MODE
  bootstrap.sh   the first-run wizard above
  plugin/        Claude commands and skills baked in
Makefile         build / build-cli / build-image / test / clean
```

## Tests

```sh
make test    # go test ./...
```
