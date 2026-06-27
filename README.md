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
git clone https://github.com/AjorCastro/agentflow
cd agentflow
go build -o agentflow ./cmd/agentflow
sudo mv agentflow /usr/local/bin/   # or any directory on your PATH
```

Or install directly:

```bash
go install github.com/agentflow/agentflow/cmd/agentflow@latest
```

### Install the CLI Agent skills

The `skills/` directory contains prompt files for Claude Code (or any compatible CLI agent).
They are embedded in the binary — install them with:

```bash
agentflow install-skills
```

This writes the skills to `~/.claude/commands/`, making them available in any project.

You can also install to a custom directory:

```bash
agentflow install-skills --target /path/to/.claude/commands
```

#### Global skills (available in any project)

These skills are installed by `agentflow install-skills` and are used before a workspace exists:

| Skill | When to use |
|---|---|
| `/agentflow-setup` | Set up a brand new project: git init, create `docs/WEB-AGENT-ROLE.md`, push to GitHub |
| `/agentflow-init`  | After the Web Reviewer creates `agentflow-init.md`: run `agentflow init` and push |

#### Project skills (installed per worktree)

These skills are copied into `.claude/commands/` inside the worktree by `agentflow init`.
They are used by the CLI Agent during feature work:

| Skill | When to use |
|---|---|
| `/agentflow-start` | First CLI turn: run discovery, write results, hand off to Web Reviewer |
| `/agentflow-plan`  | After discovery is approved: create implementation plan, hand off |
| `/agentflow-implement` | After plan is approved: execute implementation, write results |
| `/agentflow-decision` | When ambiguity or risk is found: create decision request, stop |
| `/agentflow-handoff` | At the end of any turn: update STATUS.md, state.json, commit and push |

> Project skills are coming soon.

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
  README.md                    — protocol explanation and role definitions
  STATUS.md                    — current phase, turn, status, next action
  CONFIG.md                    — feature metadata and policies
  state.json                   — machine-readable state
  specs/
  discovery/
  plans/
  tasks/
  decisions/
  reviews/
  handoffs/
  prompts/
    web-agent-role.md          — instructions for the Web Reviewer agent
```

---

## Full flow

1. **Human** describes what they want to build to the **Web Reviewer**.

2. **Web Reviewer** reads `docs/WEB-AGENT-ROLE.md`, asks clarifying questions, proposes branch name and worktree path, and creates `agentflow-init.md` in the repo root on `develop`.

3. **Human** asks the **CLI Agent** to run `/agentflow-init`. The CLI Agent reads `agentflow-init.md` and runs `agentflow init` with the parameters defined by the Web Reviewer.

4. **CLI Agent** commits and pushes. The exchange folder is now live on the feature branch, including `prompts/web-agent-role.md` with instructions for the Web Reviewer.

5. **Web Reviewer** reads the exchange folder on GitHub and begins the discovery phase.

6. **CLI Agent** executes tasks, writes results to the exchange folder, updates `STATUS.md` and `state.json`, commits and pushes.

7. **Web Reviewer / Human** reviews, approves or redirects. Repeat for each phase.

8. After merge to develop, **Human** runs `agentflow close` to clean up.

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
  skills/
    agentflow-setup.md           — /agentflow-setup global skill (new project setup)
    agentflow-init.md            — /agentflow-init global skill (workspace init)
    embed.go                     — embeds skills into the binary
  docs/
    DECISIONS.md                 — design decisions log
    WEB-AGENT-ROLE.md            — full Web Reviewer instructions (reference copy)
```
