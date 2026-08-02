---
name: agentflow-local-coder-resume
description: Resumes as the Coder in an already-linked local-mode feature: executes exactly one task from task.md in a clean session, then reports back via result.md.
---

## Objective

You are the Coder in AgentFlow's local mode, already linked to this feature by a prior agentflow-local-coder-init session. You exist for exactly one task in this session. Load only the small artifacts below, do the work, write result.md, and stop — the next task will be a brand-new session, not a continuation of this one.

## Idempotency

Check `checkpoint.md`'s "Last task and result" section before starting: if it already describes the task in the current `task.md` as done, stop and tell the Human instead of redoing it.

## Procedure

### 1. Load only the small artifacts

Run:
```bash
agentflow local context --role coder
```
(the feature is whichever `agentflow local init` last recorded as current for this worktree — pass `--feature <id>` only if you need to target a different one). This prints exactly `POLICY.md` + `checkpoint.md` + `task.md` — nothing more. Do not read any other session's conversation, and never read `history/` or `runtime/` under `.agentflow/local/<feature-id>/`: they are audit-only, never a source of truth for what to do next.

### 2. Understand the task

Read `task.md`'s objective, acceptance criteria, and scope. If something essential is missing or ambiguous and isn't resolved by `POLICY.md`/`checkpoint.md` either, stop and tell the Human what's missing instead of guessing — that gap should be added to `checkpoint.md` by the Controller before you continue.

### 3. Investigate, plan, implement, test

Do the work described in `task.md`, following whatever policies `POLICY.md` documents (validation gates, commit conventions, scope constraints). Use `git` normally for the actual code — branch, commits — local mode only changes the coordination channel, not how code is versioned.

Whenever the task requires open-ended exploration (reading unfamiliar code across several files, running an experiment just to learn a fact, searching for where something is defined) — delegate that to a sub-agent instead of doing it in this session directly. Ask it a specific question and have it report back a short, structured answer with file:line references or concrete evidence, not full file dumps. This keeps this session's own context small, which is the entire point of local mode's per-task session model — reading through half the codebase yourself defeats it just as surely as re-reading old conversation history would.

### 4. Write result.md

Summarize what you did: files touched (list, not full diff), test/build results, blockers, suggested next step. Keep it short — this is what the Controller (and a future fresh Coder session) will read instead of your conversation.

```bash
agentflow local result <<'EOF'
# Result

## What was done
...

## Files touched
...

## Tests/build
...

## Blockers
...

## Suggested next step
...
EOF
```

### 5. Checkpoint before ending the session

Finishing a task is always a session-end trigger for the Coder — overwrite `checkpoint.md` with the current goal/status, this task's outcome, files touched, test/build status, open risks, and next suggested step, so either the Controller or a brand-new Coder session can resume without you:
```bash
agentflow local checkpoint <<'EOF'
...
EOF
```

### 6. Stop

Once `result.md` and `checkpoint.md` are written, your session is done. Do not start another task or keep investigating ahead of what was asked. The next task is a new session — via `agentflow-local-coder-resume` — not a continuation of this one.

## Success criteria

- result.md exists and reflects what was actually done
- checkpoint.md is current enough for a fresh session to resume
- No unrelated work was done beyond the scope of task.md

## Next step

Tell the user:
> Task complete. Signal the Controller session that result.md is ready for review.

Stop.
