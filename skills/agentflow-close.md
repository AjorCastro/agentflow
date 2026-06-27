# AgentFlow Close

The feature has been merged. Clean up the workspace.

## Idempotency

Before executing any step, check if it was already done. If the worktree is already removed or the branch already deleted, skip those steps silently and continue. If everything is already clean, say so and stop.

## Steps

### 1. Read the current state

Read `STATUS.md` and `state.json` to get:
- `feature_branch`
- `worktree`

### 2. Confirm the feature is merged

Check with the Human that the branch has been merged to the root branch before continuing.

### 3. Run agentflow close

```bash
agentflow close --branch <feature_branch>
```

If the Human also wants to delete the remote branch:

```bash
agentflow close --branch <feature_branch> --delete-remote
```

### 4. Print summary

```
Feature  : <feature_branch>
Worktree : removed
Branch   : deleted

Next step: Run /agentflow-setup or agentflow init to start a new feature.
```

### 5. Stop

Do not start any new work.
