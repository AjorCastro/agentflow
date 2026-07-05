---
name: agentflow-init
description: Reads agentflow-init-<branch-slug>.md created by the Web Reviewer and initializes the AgentFlow workspace on a new feature branch with worktree and exchange folder. Use when the user says the Web Reviewer has created agentflow-init-<branch-slug>.md, or asks to initialize a feature, create a worktree, or run agentflow init. Trigger phrases include "the web reviewer created agentflow-init-<branch-slug>.md", "initialize the workspace", "run agentflow init", "create the feature branch".
---

# AgentFlow Init

## Core Workflow

1. Pull latest changes from origin.
2. Find and parse `agentflow-init-<branch-slug>.md` in the repository root.
3. Verify the current branch is the root branch.
4. Remove `agentflow-init-<branch-slug>.md` from root branch BEFORE init.
5. Run `agentflow init` if the workspace does not already exist.
6. Verify the result with `agentflow status`.

## Idempotency

Check each step before executing. Skip steps that are already complete. If the workspace already exists, skip init and verify the state.

## Step Detail

### Pull

The Web Reviewer created `agentflow-init-<branch-slug>.md` on GitHub. Pull first:
```
git pull --rebase
```

### Find and parse the agentflow-init file

Look for `agentflow-init-*.md` files in the current directory.
- None found after pull → stop: "No agentflow-init-*.md file found after git pull. Ask the Web Reviewer to confirm they committed it to the develop branch."
- Exactly one found → use it.
- Multiple found → list them and ask the user which feature to initialize.

Extract from the chosen file: `title`, `branch`, `worktree`, `root` (default: `develop`).

### Verify branch

Run `git rev-parse --abbrev-ref HEAD`.
- Not on root branch → stop and tell the user which branch is current.

### Remove the agentflow-init file from root branch BEFORE init

Critical: remove the file from `develop` before creating the worktree so it is not inherited by the new feature branch.

Let `<init-file>` be the filename found above.
- Exists → `git rm <init-file> && git commit -m "chore: consume <init-file>" && git push`
- Already removed → skip.

### Initialize workspace

Run `agentflow status` to check if a workspace for this branch already exists.
- Exists → skip to verification.
- Not found → run:
  ```
  agentflow init --root <root> --branch <branch> --worktree <worktree> --title "<title>"
  ```
  If no remote is configured, `agentflow init` will fail with a clear message. Tell the user to run:
  ```
  git remote add origin git@github.com:<user>/<repo>.git
  ```

### Verify result

Run `agentflow status`. Confirm:
- Exchange folder is listed
- `phase: intake`
- `turn: web`


## Constraints

- Do not begin discovery or any other task after init.
- Do not change branches in the main repo.
- The next turn belongs to the Web Reviewer.

## Success Criteria

- `agentflow status` shows exchange folder with `phase: intake`, `turn: web`
- The `agentflow-init-*.md` file no longer exists in the repo root
- Branch and worktree are pushed to GitHub

## Next Step

Tell the user:
> "Workspace ready on branch <branch>. Tell the Web Reviewer they can read the exchange folder on GitHub and begin Phase 1 — Discovery."

Stop.
