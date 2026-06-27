---
name: agentflow-setup
description: Initializes a new software project for AgentFlow collaboration. Use when the user asks to set up a new project from scratch, create a git repository, publish to GitHub, or prepare a repo so a Web Reviewer can start working with AgentFlow. Trigger phrases include "set up a new project", "initialize agentflow", "create the repo", "prepare for agentflow".
---

# AgentFlow Setup

## Core Workflow

1. Verify prerequisites (`git`, `gh` authenticated, `agentflow` installed).
2. Initialize git if not already done.
3. Create `docs/WEB-AGENT-ROLE.md` if it does not exist.
4. Create `.gitignore` with `.worktrees/` if not already present.
5. Commit any new files.
6. Create the GitHub repository and push.

## Idempotency

Check each step before executing. Skip steps that are already complete. Never duplicate work or overwrite existing files.

## Step Detail

### Prerequisites

Run and verify:
- `git --version`
- `gh auth status`
- `agentflow --help`

Stop with a clear message if any is missing.

### Git initialization

Check: `git rev-parse --git-dir`
- Not found → `git init -b develop`
- Found → confirm current branch is `develop`. If not, stop and tell the user.

### docs/WEB-AGENT-ROLE.md

Check if the file exists.
- If not → create `docs/` and write the file. Content must explain to the Web Reviewer: their role in AgentFlow, how to create `agentflow-init.md`, how to review each phase, and how to update STATUS.md. See the AgentFlow documentation at https://github.com/AjorCastro/agentflow for the canonical content.
- If exists → skip.

### .gitignore

Check if `.worktrees/` is already ignored.
- If not → create or append `.worktrees/` to `.gitignore`.
- If yes → skip.

### Commit

Check: `git status --short`
- Changes present → `git add -A && git commit -m "chore: initial project setup"`
- Clean → skip.

### GitHub repository

Check: `git remote`
- No remote → `gh repo create <project-name> --public --source . --remote origin --push`
  Ask the user if they prefer `--private`.
- Remote exists → check if branch is pushed. If not, push. If already pushed → skip.

## Constraints

- Never overwrite existing files.
- Never change the current branch without explicit user instruction.
- Ask before creating a public repository when in doubt.

## Success Criteria

- `git remote -v` shows `origin` pointing to GitHub
- `docs/WEB-AGENT-ROLE.md` exists and is committed
- `develop` branch is pushed to GitHub

## Next Step

Tell the user the GitHub URL and instruct them to share it with the Web Reviewer, asking them to read `docs/WEB-AGENT-ROLE.md`.

Stop. Do not start any feature work.
