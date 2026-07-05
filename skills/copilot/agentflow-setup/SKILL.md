---
name: agentflow-setup
description: Initializes a new software project with git, publishes it to GitHub, and creates the AgentFlow Web Reviewer instructions file. Use this when the user asks to set up a new project or start from scratch with AgentFlow.
---

## Objective

Initialize a new project repository and publish it so the Web Reviewer can start working.

## Idempotency

Before executing any step, check if it was already done. Skip completed steps silently and continue from where things left off. Never duplicate work or overwrite existing results. If everything is already done, say so and stop.

## Prerequisites

Check that the following tools are available:
- `git` — for version control
- `gh` — GitHub CLI, authenticated (`gh auth status`)
- `agentflow` — AgentFlow CLI

If any is missing, stop and tell the Human what needs to be installed before continuing.

## Procedure

### 1. Get project name

If the current directory name is not a suitable project name, ask the Human to confirm before proceeding.

Use the current directory name as the repository name unless the Human specifies otherwise.

### 2. Initialize git

Check if a git repository already exists (`git rev-parse --git-dir`).
- If it does not exist → run `git init -b develop`
- If it already exists → skip, verify current branch is `develop`. If not, tell the Human and stop.

### 3. Create docs/WEB-AGENT-ROLE.md

Run:
```bash
agentflow docs sync
```
This writes (or refreshes) `docs/WEB-AGENT-ROLE.md` on the `develop` branch directly from the canonical content built into the `agentflow` binary, and commits it if it changed.

There is nothing to write by hand here — never hand-copy this file's content into a skill file or a chat message. A hand-copied version will drift out of date the next time the binary changes; `agentflow docs sync` never can, because it always reflects whatever binary is currently installed.

### 4. Create .gitignore

Check if `.worktrees/` is already ignored.
- If not → create or append `.worktrees/` to `.gitignore`.
- If yes → skip.

### 5. Commit

Check if there are staged or unstaged changes (`git status --short`).
- If there are changes → stage and commit:
  ```bash
  git add -A
  git commit -m "chore: initial project setup"
  ```
- If the tree is already clean → skip.

### 6. Create the GitHub repository

Check if a remote named `origin` already exists (`git remote`).
- If it does not exist → create the repo and push:
  ```bash
  gh repo create <project-name> --public --source . --remote origin --push
  ```
  If the Human wants a private repository, use `--private`. When in doubt, ask.
- If `origin` already exists → check if the branch is up to date. If not, push. If already pushed → skip.

## Success criteria

- `git remote -v` shows `origin` pointing to GitHub
- `docs/WEB-AGENT-ROLE.md` exists and is committed
- `develop` branch is pushed to GitHub

## Next step

Tell the user:
> Project is ready at <github-url>. Share this URL with the Web Reviewer and ask them to read docs/WEB-AGENT-ROLE.md to start.

Stop.
