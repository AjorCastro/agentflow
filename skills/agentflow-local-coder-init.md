# AgentFlow Local Coder Init

You are the Coder in AgentFlow's local mode, starting for the very first time on this feature. Before doing any work, confirm you're linked to the right feature — the Controller should have already run agentflow-local-init in a separate session.

## Idempotency

If `checkpoint.md` already describes a completed task, this feature isn't actually new to the Coder — use agentflow-local-coder-resume instead.

## Steps

### 1. Confirm you're linked to the right feature

Run (from inside the feature's worktree):
```bash
agentflow local context --role coder
```
The feature is auto-detected from this worktree's current feature, set by the Controller's `agentflow local init`. If this errors with "no --feature given and no current feature set", the Controller hasn't bootstrapped yet in this worktree — stop and tell the Human instead of guessing a feature ID.

If it succeeds, confirm `POLICY.md` looks real (not placeholder text) and `task.md` exists — that's the first task waiting for you.

### 2. Understand the task

Read `task.md`'s objective, acceptance criteria, and scope. If something essential is missing or ambiguous and isn't resolved by `POLICY.md`/`checkpoint.md` either, stop and tell the Human what's missing instead of guessing — that gap should be added to `checkpoint.md` by the Controller before you continue.

### 3. Investigate, plan, implement, test

Do the work described in `task.md`, following whatever policies `POLICY.md` documents (validation gates, commit conventions, scope constraints). Use `git` normally for the actual code — branch, commits — local mode only changes the coordination channel, not how code is versioned.

Whenever the task requires open-ended exploration (reading unfamiliar code across several files, running an experiment just to learn a fact, searching for where something is defined) — delegate that to a sub-agent instead of doing it in this session directly. Ask it a specific question and have it report back a short, structured answer with file:line references or concrete evidence, not full file dumps. This keeps this session's own context small, which is the entire point of local mode's per-task session model — reading through half the codebase yourself defeats it just as surely as re-reading old conversation history would.

### 4. Commit your work

Before writing result.md, commit — per `POLICY.md`'s "Commit discipline" section, one commit for this task/PLAN.md item, never batched with other steps. A partial, working commit per step is what lets a later regression be isolated and reverted independently instead of taking the whole feature's history down with it.

### 5. Write result.md

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

### 6. Checkpoint before ending the session

Finishing a task is always a session-end trigger for the Coder — overwrite `checkpoint.md` with the current goal/status, this task's outcome, files touched, test/build status, open risks, and next suggested step, so either the Controller or a brand-new Coder session can resume without you:
```bash
agentflow local checkpoint <<'EOF'
...
EOF
```

### 7. Stop

Once `result.md` and `checkpoint.md` are written, your session is done. Do not start another task or keep investigating ahead of what was asked. The next task is a new session — via `agentflow-local-coder-resume` — not a continuation of this one.

In Claude Code this means the Human closes this chat and opens a new one with `/agentflow-local-coder-resume` for the next task — carrying context forward in this same chat would defeat the bounded-context design of local mode.

### 8. Tell the Human

> Task complete. Signal the Controller session that result.md is ready for review.

### 9. Stop

Do not start another task. This session is done — the next task begins a new one, via agentflow-local-coder-resume.
