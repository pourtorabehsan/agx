# CLAUDE.md

Development rules for working on `agx`. Project context lives in `README.md` and the code.

## Code

- Keep business logic in pure helpers that return values (paths, argv, slugs). Only Cobra `RunE` closures and `bootstrap.go` should touch `os.*` / `exec.Command`.
- Wrap errors with `fmt.Errorf("...: %w", err)`. On recoverable failures, add a user-actionable hint (see `bootstrap.go` and `run.go` for the pattern).
- Propagate the container's exit code from `run.go` via `os.Exit(exitErr.ExitCode())`. Don't swallow it.
- No new third-party deps without an explicit reason. Current deps: `cobra`, `golang.org/x/term`.

## Tests

- Tests live next to the code, table-driven. A new helper without a test is a smell.
- `dockerargs.go` argv shape is pinned by its tests — they encode the host/container contract. If you change the argv, update the tests in the same change.

## Image

- After editing anything under `image/plugin/`, both `make build-image` **and** `agx bootstrap` are required to land it in the user home. Rebuild alone won't do it.
- If you touch the Claude or Codex install path, relocation, or cleanup in `image/Dockerfile`, verify `claude --version` and `codex --version` run in a fresh container before claiming done.
- Pin `GO_VERSION` in `image/Dockerfile` to match `go.mod`. Bump them together.

## Git

- After a pre-commit hook fails, fix and create a **new** commit — never `--amend`.
