---
name: agentflow-close
description: Closes an AgentFlow feature workspace after the branch has been merged to the root branch. Use when the user says a feature is done, merged, or completed and asks to clean up, remove the worktree, or delete the branch. Trigger phrases include "the feature is merged", "close the feature", "clean up the workspace", "delete the branch", "agentflow close".
---

# AgentFlow Close

## Core Workflow

1. Read current state to identify the feature branch and worktree.
2. Confirm with the user that the branch is merged.
3. Run `agentflow close`.

## Idempotency

Check each step before executing. If the worktree is already removed or the branch already deleted, skip those steps. If everything is already clean, say so and stop.

## Step Detail

### Read current state

Run `agentflow status` to find `feature_branch`, `worktree`, and `root_branch`.

If no workspace is found, stop and tell the user.

### Confirm merge

Ask the user to confirm the branch has been merged before proceeding. Do not proceed without this confirmation.

### Run agentflow close

Run this from the repository checked out on `root_branch` (not from the worktree being removed):

```
agentflow close --branch <feature_branch> --root <root_branch>
```

Options:
- To also delete the remote branch: `--delete-remote`
- If the branch is not detected as merged but the user is certain: `--force`

`agentflow close` also removes `.agentflow/features/<feature-id>/` from `root_branch` if present — the merge carries those files into root, and leaving them there would clutter it permanently.

## Constraints

- Never delete a branch without explicit merge confirmation from the user.
- Never use `--force` unless the user explicitly requests it.

## Success Criteria

- `git worktree list` does not show the feature worktree
- `git branch` does not show the feature branch
- `agentflow status` shows no workspace for that feature

## Next Step

Tell the user:
> "Feature <branch> closed. Run agentflow-setup or agentflow init to start a new feature."

Stop.
