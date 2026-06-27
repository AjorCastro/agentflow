# AgentFlow

**AgentFlow** is a local CLI tool that initializes a repo-centric coordination workspace for software features involving a Human, a Web Reviewer, and a CLI Coding Agent (Claude Code, Codex, Copilot, etc.).

It is opinionated but simple: no AI, no external APIs, no GitHub integration. Just Git, worktrees, and structured Markdown files.

---

## What problem does it solve?

When working with AI coding agents on non-trivial features, coordination breaks down quickly:
- The agent implements before discovery and planning are approved.
- There is no shared state between the human, the reviewer, and the agent.
- Context is lost between sessions.

AgentFlow creates a **structured exchange folder** on the feature branch that all roles can read and write. It defines clear handoffs, phases, and policies so the agent always knows what to do next.

---

## Roles

| Role         | Responsibility                                                   |
|--------------|------------------------------------------------------------------|
| Human        | Provides requirements, approves decisions, resolves ambiguity    |
| Web Reviewer | Reviews plans and outputs via web UI; drives approval workflow   |
| CLI Agent    | Executes tasks inside the worktree; reads and writes the exchange|

**Core rule:** CLI executes. Web/Human review, decide, and approve.

---

## Install locally

Requires Go 1.22 or later.

```bash
git clone <this-repo>
cd agentflow
go build -o agentflow ./cmd/agentflow
sudo mv agentflow /usr/local/bin/   # or any directory on your PATH
```

Or install directly:

```bash
go install github.com/agentflow/agentflow/cmd/agentflow@latest
```

---

## Commands

```
agentflow init      Initialize a new AgentFlow workspace
agentflow status    Show current workspace status
agentflow validate  Validate workspace structure
agentflow close     Close a feature after merge
```

---

## Example: `agentflow init`

```bash
agentflow init \
  --repo . \
  --root develop \
  --branch feature/my-feature \
  --worktree .worktrees/my-feature \
  --exchange .agentflow/features/my-feature \
  --title "My Feature" \
  --push
```

### Flags

| Flag            | Default                              | Description                                     |
|-----------------|--------------------------------------|-------------------------------------------------|
| `--repo`        | `.`                                  | Path to the git repository                      |
| `--root`        | `develop`                            | Root branch to branch from                      |
| `--branch`      | *(required)*                         | Feature branch name                             |
| `--worktree`    | *(required unless --no-worktree)*    | Worktree directory path                         |
| `--exchange`    | `.agentflow/features/<feature-id>`   | Exchange folder path                            |
| `--title`       | *(feature-id)*                       | Human-readable feature title                    |
| `--push`        | `true`                               | Push branch to origin after init                |
| `--adopt`       | `false`                              | Attach worktree to an existing branch           |
| `--no-worktree` | `false`                              | Skip worktree; operate on current repo branch   |

### What it creates

```
.agentflow/features/my-feature/
  README.md       — protocol explanation and role definitions
  STATUS.md       — current phase, turn, status, next action
  CONFIG.md       — feature metadata and policies
  state.json      — machine-readable state
  specs/
  discovery/
  plans/
  tasks/
  decisions/
  reviews/
  handoffs/
  prompts/
```

---

## Full flow

1. **Human + Web Reviewer** agree on feature scope, branch name, worktree path.

2. **Human** runs `agentflow init` (see above).

3. **Human** opens the worktree in the CLI agent (Claude Code, etc.) and asks it to run the initial AgentFlow skill.

4. **CLI Agent** reads `STATUS.md` and `CONFIG.md`, executes discovery, writes results to `discovery/`, updates `STATUS.md` and `state.json`, hands off.

5. **Web Reviewer / Human** reviews discovery output, approves or redirects.

6. Repeat for planning, implementation, review, etc.

7. After merge to develop, **Human** runs `agentflow close` to clean up.

---

## Example: `agentflow close`

```bash
agentflow close --branch feature/my-feature
```

Verifies the branch is merged, removes the linked worktree, and deletes the local branch.

### Flags

| Flag              | Default  | Description                              |
|-------------------|----------|------------------------------------------|
| `--repo`          | `.`      | Path to the git repository               |
| `--branch`        | *(required)* | Feature branch to close              |
| `--delete-remote` | `false`  | Also delete the branch on origin         |
| `--force`         | `false`  | Close even if branch is not fully merged |

---

## Running tests

```bash
go test ./...
```

---

## Project structure

```
agentflow/
  cmd/agentflow/main.go          — entry point
  internal/cli/
    root.go                      — cobra root command
    init.go                      — agentflow init
    status.go                    — agentflow status
    validate.go                  — agentflow validate
    close.go                     — agentflow close
  internal/gitops/
    git.go                       — git operations via os/exec
  internal/protocol/
    state.go                     — state.json read/write
    templates.go                 — file content generators
    validate.go                  — structure validation, path helpers
  docs/
    DECISIONS.md                 — design decisions log
```
