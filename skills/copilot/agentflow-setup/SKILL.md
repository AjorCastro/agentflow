---
name: agentflow-setup
description: Initializes a new software project with git, publishes it to GitHub, and creates the AgentFlow Web Reviewer instructions file. Use this when the user asks to set up a new project or start from scratch with AgentFlow.
---

## Objective

Set up a brand new project repository so the Web Reviewer can start collaborating via GitHub.

## Idempotency

Before executing any step, check if it was already done. Skip completed steps silently. If everything is already done, say so and stop.

## Prerequisites

Verify these tools are available before starting:
- `git` — version control
- `gh` — GitHub CLI, authenticated (`gh auth status`)
- `agentflow` — AgentFlow CLI (`agentflow --help`)

If any is missing, stop and tell the user what needs to be installed.

## Procedure

### 1. Initialize git

Check if a git repository already exists (`git rev-parse --git-dir`).
- If it does not exist → run `git init -b develop`
- If it already exists → verify current branch is `develop`. If not, stop and tell the user.

### 2. Create docs/WEB-AGENT-ROLE.md

Check if `docs/WEB-AGENT-ROLE.md` already exists.
- If it does not exist → create the `docs/` directory and copy the content from `prompts/web-agent-role.md` in the AgentFlow exchange folder, or use the content embedded in the `agentflow` binary (run `agentflow install-skills` to see available content).

The file must explain to the Web Reviewer:
- Their role in the AgentFlow protocol
- How to create `agentflow-init.md`
- How to review discovery, plans, and implementation
- How to update STATUS.md

### 3. Create .gitignore

Check if `.gitignore` exists and contains `.worktrees/`.
- If not → create or append `.worktrees/` to `.gitignore`.

### 4. Commit

Check if there are uncommitted changes (`git status --short`).
- If yes → stage and commit: `git add -A && git commit -m "chore: initial project setup"`
- If already clean → skip.

### 5. Create GitHub repository and push

Check if a remote named `origin` already exists (`git remote`).
- If not → run: `gh repo create <project-name> --public --source . --remote origin --push`
  Ask the user if they prefer `--private` when in doubt.
- If already exists → check if branch is pushed. If not, push. If already pushed → skip.

## Success criteria

- `git remote -v` shows `origin` pointing to a GitHub URL
- `docs/WEB-AGENT-ROLE.md` exists and is committed
- The branch `develop` is pushed to GitHub

## Next step

Tell the user:
> "Project is ready at <github-url>. Share this URL with the Web Reviewer and ask them to read docs/WEB-AGENT-ROLE.md to start."
