---
name: agentflow-init
description: Reads agentflow-init.md created by the Web Reviewer and initializes the AgentFlow workspace. Use this when the user says the Web Reviewer has created agentflow-init.md and asks you to initialize the workspace.
---

## Objective

Read the parameters defined by the Web Reviewer and run `agentflow init` to create the feature branch, worktree, and exchange folder.

## Idempotency

Before executing any step, check if it was already done. Skip completed steps silently. If everything is already done, say so and stop.

## Procedure

### 1. Find and read agentflow-init.md

Look for `agentflow-init.md` in the repository root.

If it does not exist, stop and tell the user:
> "agentflow-init.md not found. Ask the Web Reviewer to create it first."

Extract these parameters:
- `title`
- `branch`
- `worktree`
- `root` (default: `develop` if not specified)

### 2. Verify current branch

Run `git rev-parse --abbrev-ref HEAD`.
If you are not on the root branch, stop and tell the user.

### 3. Initialize the workspace

Run `agentflow status` to check if a workspace for this branch already exists.
- If it exists → skip to step 4.
- If not → run:

```
agentflow init \
  --root <root> \
  --branch <branch> \
  --worktree <worktree> \
  --title "<title>"
```

The `--push` flag is enabled by default. If there is no remote, `agentflow init` will fail with a clear message — tell the user to add a remote first.

### 4. Verify the result

Run `agentflow status` and confirm:
- Exchange folder exists
- `phase` is `intake`
- `turn` is `human`

### 5. Remove agentflow-init.md

Check if `agentflow-init.md` still exists.
- If yes → remove and commit:
  ```
  git rm agentflow-init.md
  git -C <worktree> add -A
  git -C <worktree> commit -m "chore: remove agentflow-init.md after workspace setup"
  git -C <worktree> push
  ```
- If already removed → skip.

## Success criteria

- `agentflow status` shows the exchange folder with `phase: intake`, `turn: human`
- `agentflow-init.md` no longer exists in the repo root
- Branch and worktree are pushed to GitHub

## Next step

Tell the user:
> "Workspace ready on branch <branch>. Tell the Web Reviewer they can now read the exchange folder on GitHub and begin Phase 1 — Discovery."

Stop. Do not begin any discovery or other task.
