# AgentFlow Local Pause

The Human needs to stop this session now, but you're not at one of the normal restart/end triggers (task done, plan approved, deliberation closed). Do not just end the conversation — write down what's true right now so the next session isn't starting blind.

## Idempotency

If you're already at a normal milestone (task just finished, plan just approved), use the ordinary checkpoint step from your role's skill instead — this skill is only for stopping mid-work.

## Steps

### 1. Do not reload context

You're pausing an active session, not starting one — everything you need is already in this conversation. Do not re-run `agentflow local context`.

### 2. Write a partial checkpoint.md

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

### 3. If you are the Coder and the task is genuinely incomplete, say so

Do not write `result.md` for an unfinished task — result.md means the task is done. The checkpoint alone is what carries a paused, in-progress task forward.

### 4. Tell the Human it's safe to close this session

Confirm the checkpoint was written, then stop. The Human resumes later via agentflow-local-resume (Controller) or agentflow-local-coder-resume (Coder) — never a continuation of this session.

### 5. Tell the Human

> Paused. Resume later with agentflow-local-resume or agentflow-local-coder-resume, depending on which role this was.

### 6. Stop

Session ends here. Do not continue working after writing the pause checkpoint.
