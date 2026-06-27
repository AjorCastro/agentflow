---
name: agentflow-init
description: Reads agentflow-init.md created by the Web Reviewer and initializes the AgentFlow workspace on a new feature branch with worktree and exchange folder. Use when the user says the Web Reviewer has created agentflow-init.md, or asks to initialize a feature, create a worktree, or run agentflow init. Trigger phrases include "the web reviewer created agentflow-init.md", "initialize the workspace", "run agentflow init", "create the feature branch".
---

# AgentFlow Init

## Core Workflow

1. Find and parse `agentflow-init.md` in the repository root.
2. Verify the current branch is the root branch.
3. Run `agentflow init` if the workspace does not already exist.
4. Verify the result with `agentflow status`.
5. Remove `agentflow-init.md` and push.

## Idempotency

Check each step before executing. Skip steps that are already complete. If the workspace already exists, skip init and verify the state.

## Step Detail

### Parse agentflow-init.md

Look for `agentflow-init.md` in the current directory.
- Not found → stop: "agentflow-init.md not found. Ask the Web Reviewer to create it first."
- Found → extract: `title`, `branch`, `worktree`, `root` (default: `develop`).

### Verify branch

Run `git rev-parse --abbrev-ref HEAD`.
- Not on root branch → stop and tell the user which branch is current.

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
- `turn: human`

### Remove agentflow-init.md

Check if the file still exists.
- Exists → remove and commit:
  ```
  git rm agentflow-init.md
  git -C <worktree> add -A
  git -C <worktree> commit -m "chore: remove agentflow-init.md after workspace setup"
  git -C <worktree> push
  ```
- Already removed → skip.

## Constraints

- Do not begin discovery or any other task after init.
- Do not change branches in the main repo.
- The next turn belongs to the Web Reviewer.

## Success Criteria

- `agentflow status` shows exchange folder with `phase: intake`, `turn: human`
- `agentflow-init.md` no longer exists in the repo root
- Branch and worktree are pushed to GitHub

## Next Step

Tell the user:
> "Workspace ready on branch <branch>. Tell the Web Reviewer they can read the exchange folder on GitHub and begin Phase 1 — Discovery."

Stop.
