# AgentFlow Config

feature_id: feature-test-init
title: Test init local
root_branch: develop
feature_branch: feature/test-init
worktree: /home/ajc/projects/agentflow/.worktrees/test-init
exchange_folder: /home/ajc/projects/agentflow/.worktrees/test-init/.agentflow/features/feature-test-init

## Policies

- CLI must not implement before discovery and planning are approved.
- CLI must create result files at the end of each task.
- CLI must create decision requests when ambiguity or risk is found.
- CLI must update STATUS.md and state.json at every handoff.
- CLI may commit and push at the end of a completed turn.
