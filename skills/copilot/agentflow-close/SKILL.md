---
name: agentflow-close
description: Closes an AgentFlow feature workspace after the branch has been merged. Use this when the user says a feature has been merged and asks to clean up the workspace.
---

## Objective

Remove the worktree and delete the feature branch after a successful merge to the root branch.

## Idempotency

Before executing any step, check if it was already done. If the worktree is already removed or the branch already deleted, skip those steps silently. If everything is already clean, say so and stop.

## Procedure

### 1. Read current state

Run `agentflow status` to find the `feature_branch` and `worktree` values.

If no AgentFlow workspace is found, stop and tell the user.

### 2. Confirm the merge

Ask the user to confirm that the branch has been merged to the root branch before proceeding.

Do not proceed without this confirmation.

### 3. Run agentflow close

```
agentflow close --branch <feature_branch>
```

If the user also wants to delete the remote branch:
```
agentflow close --branch <feature_branch> --delete-remote
```

If the branch is not detected as merged and the user is certain it was merged, use `--force`:
```
agentflow close --branch <feature_branch> --force
```

## Success criteria

- `git worktree list` no longer shows the feature worktree
- `git branch` no longer shows the feature branch
- `agentflow status` shows no workspace for that feature

## Next step

Tell the user:
> "Feature <branch> closed. Run /agentflow-setup or agentflow init to start a new feature."

Stop.
