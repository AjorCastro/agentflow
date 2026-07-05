# AgentFlow Setup

Initialize a new project repository and publish it so the Web Reviewer can start working.

## Idempotency

Before executing any step, check if it was already done. Skip completed steps silently and continue from where things left off. Never duplicate work or overwrite existing results. If everything is already done, say so and stop.

## Steps

### 1. Verify prerequisites

Check that the following tools are available:
- `git` — for version control
- `gh` — GitHub CLI, authenticated (`gh auth status`)
- `agentflow` — AgentFlow CLI

If any is missing, stop and tell the Human what needs to be installed before continuing.

### 2. Get project name

If the current directory name is not a suitable project name, ask the Human to confirm before proceeding.

Use the current directory name as the repository name unless the Human specifies otherwise.

### 3. Initialize git

Check if a git repository already exists (`git rev-parse --git-dir`).
- If it does not exist → run `git init -b develop`
- If it already exists → skip, verify current branch is `develop`. If not, tell the Human and stop.

### 4. Create docs/WEB-AGENT-ROLE.md

Run:
```bash
agentflow docs sync
```
This writes (or refreshes) `docs/WEB-AGENT-ROLE.md` on the `develop` branch directly from the canonical content built into the `agentflow` binary, and commits it if it changed.

There is nothing to write by hand here — never hand-copy this file's content into a skill file or a chat message. A hand-copied version will drift out of date the next time the binary changes; `agentflow docs sync` never can, because it always reflects whatever binary is currently installed.

### 5. Create .gitignore

Check if `.worktrees/` is already ignored.
- If not → create or append `.worktrees/` to `.gitignore`.
- If yes → skip.

### 6. Commit

Check if there are staged or unstaged changes (`git status --short`).
- If there are changes → stage and commit:
  ```bash
  git add -A
  git commit -m "chore: initial project setup"
  ```
- If the tree is already clean → skip.

### 7. Create the GitHub repository

Check if a remote named `origin` already exists (`git remote`).
- If it does not exist → create the repo and push:
  ```bash
  gh repo create <project-name> --public --source . --remote origin --push
  ```
  If the Human wants a private repository, use `--private`. When in doubt, ask.
- If `origin` already exists → check if the branch is up to date. If not, push. If already pushed → skip.

### 8. Print summary

```
Project   : <project-name>
Branch    : develop
Remote    : <github-url>
Next step : Share the GitHub URL with the Web Reviewer and ask them to read docs/WEB-AGENT-ROLE.md
```

### 9. Stop

Do not continue with any other task. Wait for the Human to coordinate with the Web Reviewer.
