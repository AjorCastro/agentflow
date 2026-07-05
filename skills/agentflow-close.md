# AgentFlow Close

The feature has been merged. Clean up the workspace.

## Idempotency

Before executing any step, check if it was already done. If the worktree is already removed or the branch already deleted, skip those steps silently and continue. If everything is already clean, say so and stop.

## Steps

### 1. Read the current state

Run `agentflow status` to find `feature_branch`, `worktree`, and `root_branch`.

If no AgentFlow workspace is found, stop and tell the Human.

### 2. Confirm the feature is merged

Ask the Human to confirm that the branch has been merged to the root branch before continuing. Do not proceed without this confirmation.

### 3. Run agentflow close

Run this from the repository checked out on `root_branch` (not from the worktree being removed):

```bash
agentflow close --branch <feature_branch> --root <root_branch>
```

If the Human also wants to delete the remote branch:
```bash
agentflow close --branch <feature_branch> --root <root_branch> --delete-remote
```

If the branch is not detected as merged but the Human is certain it was merged, use `--force` — but never use it unless the Human explicitly asks for it:
```bash
agentflow close --branch <feature_branch> --root <root_branch> --force
```

`agentflow close` also removes `.agentflow/features/<feature-id>/` from `root_branch` if present — the merge carries those files into root, and leaving them there would clutter it permanently.

### 4. Print summary

```
Feature  : <feature_branch>
Worktree : removed
Branch   : deleted
Exchange : removed from <root_branch>

Next step: Run /agentflow-setup or agentflow init to start a new feature.
```

### 5. Stop

Do not start any new work.
