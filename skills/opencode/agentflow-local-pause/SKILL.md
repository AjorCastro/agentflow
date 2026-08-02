---
name: agentflow-local-pause
description: Safely stops the current Controller or Coder session in AgentFlow's local mode before a natural milestone, by writing a partial checkpoint.md that captures in-progress work, so the next agentflow-local-resume or agentflow-local-coder-resume doesn't start blind. Use when the Human needs to interrupt a session for a reason unrelated to task/plan completion — end of day, an external interruption. Trigger phrases include "pause here", "agentflow local pause", "I need to stop for now", "let's continue later".
---

# AgentFlow Local Pause

## Core Workflow

1. Do not reload context
2. Write a partial checkpoint.md
3. If you are the Coder and the task is genuinely incomplete, say so
4. Tell the Human it's safe to close this session

## Idempotency

If you're already at a normal milestone (task just finished, plan just approved), use the ordinary checkpoint step from your role's skill instead — this skill is only for stopping mid-work.

## Step Detail

### Do not reload context

You're pausing an active session, not starting one — everything you need is already in this conversation. Do not re-run `agentflow local context`.

### Write a partial checkpoint.md

Overwrite `checkpoint.md`, explicitly marked as a mid-work pause (not a completed milestone), covering: overall goal/status, exactly what's done and what's still in progress in the current task/round, files touched so far, test/build status if known, open blockers, and the precise next step — specific enough that a resumed session doesn't have to guess or redo work you already did.
```bash
agentflow local checkpoint <<'EOF'
# Checkpoint

## Goal and overall status
... (note: PAUSED mid-work, not a natural milestone) ...

## Last task and result
... (in progress: what's done, what's left) ...

## Files touched
...

## Test/build status
...

## Open blockers or risks
...

## Next suggested step
...
EOF
```

### If you are the Coder and the task is genuinely incomplete, say so

Do not write `result.md` for an unfinished task — result.md means the task is done. The checkpoint alone is what carries a paused, in-progress task forward.

### Tell the Human it's safe to close this session

Confirm the checkpoint was written, then stop. The Human resumes later via agentflow-local-resume (Controller) or agentflow-local-coder-resume (Coder) — never a continuation of this session.

## Constraints

- Never end a mid-work session without writing a partial checkpoint first.
- Never write result.md for an incomplete task just to have something to show.
- Never treat a pause as a normal milestone checkpoint — say explicitly in checkpoint.md that this was a pause, not a completion.

## Success Criteria

- checkpoint.md accurately reflects in-progress, incomplete work — not a finished round
- A fresh resume session could pick up exactly where this one stopped, without re-asking the Human what was happening

## Next Step

Tell the user:
> Paused. Resume later with agentflow-local-resume or agentflow-local-coder-resume, depending on which role this was.

Stop.
