---
name: agentflow-init
description: Reads agentflow-init.md created by the Web Reviewer and initializes the AgentFlow workspace. Use this when the user says the Web Reviewer has created agentflow-init.md and asks you to initialize the workspace.
---

## Objective

Read the parameters defined by the Web Reviewer and run `agentflow init` to create the feature branch, worktree, and exchange folder.

## Idempotency

Before executing any step, check if it was already done. Skip completed steps silently. If everything is already done, say so and stop.

## Procedure

### 1. Pull latest changes

The Web Reviewer created `agentflow-init.md` on GitHub. Pull it first:
```
git pull --rebase
```

### 2. Find and read the agentflow-init file

Look for `agentflow-init-*.md` files in the repository root.

- If no file is found after pulling, stop and tell the user:
  > "No agentflow-init-*.md file found after git pull. Ask the Web Reviewer to confirm they committed it to the develop branch."
- If exactly one file is found, use it.
- If multiple files are found, list them and ask the user which feature to initialize.

Extract these parameters:
- `title`
- `branch`
- `worktree`
- `root` (default: `develop` if not specified)

### 3. Verify current branch

Run `git rev-parse --abbrev-ref HEAD`.
If you are not on the root branch, stop and tell the user.

### 4. Remove agentflow-init.md from root branch BEFORE init

Critical: remove the file from `develop` before creating the worktree so it is not inherited by the new feature branch.

Let `<init-file>` be the filename found in step 2.

Check if it still exists:
- If yes → remove and commit on the root branch:
  ```
  git rm <init-file>
  git commit -m "chore: consume <init-file>"
  git push
  ```
- If already removed → skip.

### 5. Initialize the workspace

Run `agentflow status` to check if a workspace for this branch already exists.
- If it exists → skip to step 6.
- If not → run:

```
agentflow init \
  --root <root> \
  --branch <branch> \
  --worktree <worktree> \
  --title "<title>"
```

The `--push` flag is enabled by default. If there is no remote, `agentflow init` will fail with a clear message — tell the user to add a remote first.

### 6. Verify the result

Run `agentflow status` and confirm:
- Exchange folder exists
- `phase` is `intake`
- `turn` is `web`

## Success criteria

- `agentflow status` shows the exchange folder with `phase: intake`, `turn: web`
- The `agentflow-init-*.md` file no longer exists on `develop`
- Branch and worktree are pushed to GitHub

## Next step

Tell the user:
> "Workspace ready on branch <branch>. Tell the Web Reviewer they can now read the exchange folder on GitHub and begin Phase 1 — Discovery."

Stop. Do not begin any discovery or other task.
