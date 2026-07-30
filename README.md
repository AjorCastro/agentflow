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

## Quickstart

### Step 1 — Install

```bash
# Install the CLI
go install github.com/AjorCastro/agentflow/cmd/agentflow@latest

# Install the skills globally
agentflow install-skills                  # Claude Code (default)
agentflow install-skills --agent copilot  # Copilot CLI
agentflow install-skills --agent codex    # Codex CLI
agentflow install-skills --agent kimi     # Kimi Code CLI
```

### Step 2 — Set up a new project

Create a new project folder, open it in your CLI Agent (Claude Code, Copilot, Codex, Kimi Code), and run:

```
/agentflow-setup
```

The CLI Agent will:
- Initialize git on `develop`
- Create `docs/WEB-AGENT-ROLE.md` with instructions for the Web Reviewer
- Create the GitHub repository and push

### Step 3 — Brief the Web Reviewer

Share the GitHub repository URL with your Web Reviewer (ChatGPT, Claude.ai, Gemini, etc.) and tell them:

> "Read `docs/WEB-AGENT-ROLE.md` — it explains your role. Then let's define the first feature."

### Step 4 — Initialize a feature

The Web Reviewer will ask clarifying questions, propose the branch/worktree names, and create `agentflow-init-<branch-slug>.md` in the repo root (e.g. `agentflow-init-feature-my-feature.md`). Then ask your CLI Agent to run:

```
/agentflow-init
```

> **Naming note:** `branch-slug` and `feature-id` are two different derived names — don't confuse them. `branch-slug` is the full branch name with every `/` turned into `-` (used only for the init request filename above). `feature-id` is the same, but with a leading `feature/` dropped first (used for the exchange folder, since `.agentflow/features/` already says "feature"). For branch `feature/my-feature`: `branch-slug` = `feature-my-feature`, `feature-id` = `my-feature` → folder `.agentflow/features/my-feature/`.

### Step 5 — Work

From this point, the cycle is:
- CLI Agent's turn → `/agentflow-turn`
- Web Reviewer reviews on GitHub, approves or redirects
- Repeat until the feature is done

When the Web Reviewer and Human are satisfied, the **Human** merges the feature branch into the root branch (e.g. via a pull request, or `git merge`/`git push` directly) — AgentFlow does not do this automatically. Only after that merge is complete should you move to Step 6.

### Step 6 — Close

After merging to `develop`, ask the CLI Agent to run:

```
/agentflow-close
```

This removes the worktree, deletes the feature branch, and also removes `.agentflow/features/<feature-id>/` from `develop` — the merge in Step 5 carried those files into `develop` along with the rest of the feature's changes, so `close` cleans them up as its last step.

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
go install github.com/AjorCastro/agentflow/cmd/agentflow@latest
```

### Install the CLI Agent skills

The `skills/` directory contains prompt files for Claude Code (or any compatible CLI agent).
They are embedded in the binary — install them with:

```bash
# Claude Code (default)
agentflow install-skills

# GitHub Copilot CLI
agentflow install-skills --agent copilot

# OpenAI Codex CLI
agentflow install-skills --agent codex

# Kimi Code CLI
agentflow install-skills --agent kimi
```

| Agent | Install location | Activation |
|---|---|---|
| Claude Code | `~/.claude/commands/` | Available as `/agentflow-*` slash commands |
| Copilot CLI | `~/.copilot/skills/` | Run `/skills reload` inside Copilot |
| Codex CLI | `~/.codex/skills/` | Auto-detected from `description` field |
| Kimi Code CLI | `~/.kimi-code/skills/` (or `$KIMI_CODE_HOME/skills/`) | Run `/reload` or start a new session, then invoke with `/skill:<name>`, e.g. `/skill:agentflow-turn` |

#### Skills (all agents)

| Skill | When to use |
|---|---|
| `agentflow-setup` | Set up a brand new project: git init, create `docs/WEB-AGENT-ROLE.md`, push to GitHub |
| `agentflow-init`  | After the Web Reviewer creates `agentflow-init-<branch>.md`: run `agentflow init` and push |
| `agentflow-turn`  | Any CLI Agent turn: read STATUS.md, act on current phase, commit and push |
| `agentflow-close` | After merge: remove worktree and delete branch |


---

## Commands

```
agentflow init        Initialize a new AgentFlow workspace
agentflow status      Show current workspace status
agentflow validate    Validate workspace structure
agentflow close       Close a feature after merge
agentflow docs sync   Write/refresh docs/WEB-AGENT-ROLE.md from the binary's canonical content
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
    PLAN.md                    — the one current plan; revised in place, with a Changelog section
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

2. **Web Reviewer** reads `docs/WEB-AGENT-ROLE.md`, asks clarifying questions, proposes branch name and worktree path, and creates `agentflow-init-<branch-slug>.md` in the repo root on `develop` (e.g. `agentflow-init-feature-my-feature.md`). Using a branch-specific filename allows multiple features to be initialized concurrently without collision.

3. **Human** asks the **CLI Agent** to run `/agentflow-init`. The CLI Agent finds the `agentflow-init-*.md` file and runs `agentflow init` with the parameters defined by the Web Reviewer.

4. **CLI Agent** commits and pushes. The exchange folder is now live on the feature branch, including `prompts/web-agent-role.md` with instructions for the Web Reviewer.

5. **Web Reviewer** reads the exchange folder on GitHub and begins the discovery phase.

6. **CLI Agent** executes tasks, writes results to the exchange folder, updates `STATUS.md` and `state.json`, commits and pushes.

7. **Web Reviewer / Human** reviews, approves or redirects. Repeat for each phase.

8. After merge to develop, **Human** runs `agentflow close` to clean up — removes the worktree, deletes the branch, and removes the now-merged `.agentflow/features/<feature-id>/` folder from develop.

---

## Example: `agentflow close`

```bash
agentflow close --branch feature/my-feature --root develop
```

Verifies the branch is merged, removes the linked worktree, deletes the local branch, and removes `.agentflow/features/<feature-id>/` from `--root` if present (the merge carries it there along with the rest of the feature's changes). Must be run from the repository checked out on `--root`, not from the worktree being removed.

### Flags

| Flag              | Default  | Description                              |
|-------------------|----------|------------------------------------------|
| `--repo`          | `.`      | Path to the git repository               |
| `--root`          | `develop` | Root branch the feature was merged into |
| `--branch`        | *(required)* | Feature branch to close              |
| `--delete-remote` | `false`  | Also delete the branch on origin         |
| `--force`         | `false`  | Close even if branch is not fully merged |

---

## Running tests

```bash
go test ./...
```

## Editing skills

The 12 files under `skills/` (4 skills × Claude/Codex/Copilot) are generated
from `internal/skillgen`, never hand-edited. To change a skill's steps, edit
its `SkillDef` in `internal/skillgen/skills_*.go`, then:

```bash
go generate ./...
go test ./...     # fails if skills/ doesn't match the SkillDefs
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
    docs.go                      — agentflow docs sync
    install_skills.go            — agentflow install-skills
  internal/gitops/
    git.go                       — git operations via os/exec
  internal/protocol/
    state.go                     — state.json read/write
    templates.go                 — file content generators (incl. WebAgentRoleMD)
    validate.go                  — structure validation, path helpers
  internal/skillgen/
    skillgen.go                  — SkillDef/Step/Flavor types, Render dispatch
    render_{claude,codex,copilot}.go — per-flavor markdown structure (Kimi reuses render_codex.go — identical SKILL.md convention)
    skills_{init,setup,turn,close}.go — the 4 skills' SkillDefs (single source)
    skillgen_test.go             — fails if skills/ is stale vs. these SkillDefs
  cmd/gen-skills/main.go         — writes skills/ from internal/skillgen
  skills/
    agentflow-*.md               — Claude Code skills (generated, see internal/skillgen)
    copilot/agentflow-*/SKILL.md — Copilot CLI skills (generated)
    codex/agentflow-*/SKILL.md   — Codex CLI skills (generated)
    kimi/agentflow-*/SKILL.md    — Kimi Code CLI skills (generated)
    embed.go                     — embeds all skills into the binary; go:generate lives here
  docs/
    DECISIONS.md                 — design decisions log
    WEB-AGENT-ROLE.md            — Web Reviewer instructions (kept in sync with
                                    protocol.WebAgentRoleMD() via `agentflow docs sync`)
```
