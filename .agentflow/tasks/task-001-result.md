# Task 001 — Result: AgentFlow MVP CLI

**Date:** 2026-06-27  
**Turn:** CLI Agent (initial)  
**Phase:** intake → implementation (completed)

---

## Files Created

```
agentflow/
  README.md
  go.mod / go.sum
  cmd/agentflow/main.go
  internal/cli/root.go
  internal/cli/init.go
  internal/cli/status.go
  internal/cli/validate.go
  internal/gitops/git.go
  internal/protocol/state.go
  internal/protocol/state_test.go
  internal/protocol/templates.go
  internal/protocol/validate.go
  internal/protocol/validate_test.go
  docs/DECISIONS.md
  .agentflow/tasks/task-001-result.md   ← this file
  .agentflow/STATUS.md
```

---

## Commands Executed

```bash
# Installed Go 1.24.4
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# Initialized module
go mod init github.com/agentflow/agentflow
go get github.com/spf13/cobra@latest

# Built binary
go build -o /tmp/agentflow ./cmd/agentflow

# Ran tests
go test ./...      # 7/7 PASS

# End-to-end test with real Git
agentflow init \
  --repo <test-repo> \
  --root develop \
  --branch feature/test-agentflow \
  --worktree <test-worktree> \
  --title "Test AgentFlow Init"

agentflow status --repo <test-worktree>
agentflow validate --repo <test-worktree>

# Initialized project as git repo
git init -b develop
git add -A && git commit -m "feat: initial AgentFlow CLI implementation"
```

---

## Decisions Taken

| ID  | Decision | Rationale |
|-----|----------|-----------|
| 001 | `os/exec` for Git, not go-git | Worktree support, zero deps, real git error messages |
| 002 | Exchange defaults to `.agentflow/features/<feature-id>/` | Discoverable via glob, namespaced, derives from branch |
| 003 | `--worktree` required; `--no-worktree` as escape hatch | No implicit checkout; fail-safe for destructive ops |
| 004 | No TUI, no interactive prompts | Composable with scripts; TUI can come later |
| 005 | Paths in state.json stored as provided | Portability tradeoff; marked as open in DECISIONS.md |

Full details in `docs/DECISIONS.md`.

---

## Bug Found and Fixed During Testing

**Exchange path created in repo root instead of worktree.**

- Root cause: `opts.exchange` default was computed using `repoAbs` before `workDir` was determined.
- Fix: moved default exchange path resolution to after `workDir` is set.
- The `--exchange` flag (user-provided) was also only resolved as absolute after `workDir` is known.

---

## Test Results

```
ok  github.com/agentflow/agentflow/internal/protocol    0.004s
--- PASS: TestNewInitialState
--- PASS: TestWriteAndReadStateFile
--- PASS: TestFeatureIDFromBranch
--- PASS: TestDefaultExchangePath
--- PASS: TestValidateExchange_Valid
--- PASS: TestValidateExchange_MissingFiles
--- PASS: TestValidateExchange_MissingOneDir
```

End-to-end: `init`, `status`, `validate` all working correctly against a real Git repo.

---

## Open Questions / Doubts

1. **Relative vs absolute paths in state.json** — Currently stores whatever the user passed. Normalizing to relative-from-repo would improve portability but requires passing repo root through more of the code. Marked as DECISION-005 (open).

2. **`--no-worktree` with implicit branch switch** — Currently fails if HEAD is not on the target branch. Should we add a `--checkout` flag to allow safe switch? Deferred to review.

3. **Go module path** — Using `github.com/agentflow/agentflow`. If this repo lives elsewhere, the module path should be updated. No functional impact until publishing.

4. **`git add -A` in CommitAll** — Uses `git add -A` which stages everything in the worktree. This is fine for the init commit, but future use in a non-empty worktree should be more selective. Consider passing explicit paths.

5. **`--push` flag with no remote** — Currently fails with git error. Could be caught earlier with a better message. Deferred.

---

## How to Test the CLI

```bash
# Install
export PATH=$PATH:/usr/local/go/bin
go build -o agentflow ./cmd/agentflow

# Show help
./agentflow --help
./agentflow init --help

# Full test with a real repo
git init -b develop /tmp/test-repo
git -C /tmp/test-repo config user.email "you@example.com"
git -C /tmp/test-repo config user.name "Test"
echo "hello" > /tmp/test-repo/hello.txt
git -C /tmp/test-repo add . && git -C /tmp/test-repo commit -m "init"

./agentflow init \
  --repo /tmp/test-repo \
  --root develop \
  --branch feature/my-feature \
  --worktree /tmp/test-wt \
  --title "My Feature"

./agentflow status --repo /tmp/test-wt
./agentflow validate --repo /tmp/test-wt

# Run unit tests
go test ./...
```

---

## Recommended Next Step

**Web Reviewer / Human review this turn before proceeding.**

Specific items to review:

1. Confirm the file structure and content of generated files matches expectations.
2. Confirm the `state.json` schema (especially field names, `last_result: null`, etc.).
3. Decide on open question #1 (relative paths in state.json).
4. Decide on open question #2 (--no-worktree + implicit checkout).
5. If approved, next turn: implement the initial AgentFlow CLI skill (the skill that reads this exchange folder and bootstraps discovery).
