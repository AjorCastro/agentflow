---
name: agentflow-local-resume
description: Resumes as the Controller in AgentFlow's local mode, in a feature already bootstrapped by agentflow-local-init: talks to the Human, defines the Coder's next task, reviews its results, and manages isolated human deliberations. Never implements code. Use for every Controller session after the first one — including forced session restarts mid-feature. Trigger phrases include "resume as Controller", "agentflow local resume", "assign the next task", "review the Coder's result", "let's deliberate".
---

# AgentFlow Local Resume (Controller)

## Core Workflow

1. Load only the small artifacts
2. Converse with the Human and decide the next step
3. Define the next task
4. Review the Coder's result
5. Checkpoint at a milestone
6. Isolated human deliberations
7. Restart your session at the right triggers

## Idempotency

Before doing anything else, check whether `task.md` or `result.md` already exist for the current round via `agentflow local context --role controller`. If a `result.md` is already waiting for review, review it before defining a new task — do not overwrite `task.md` mid-round.

## Step Detail

### Load only the small artifacts

Run:
```bash
agentflow local context --role controller
```
(the feature is whichever `agentflow local init` last recorded as current for this worktree — pass `--feature <id>` only if you need to target a different one). This prints exactly `POLICY.md` + `checkpoint.md` + `task.md` + `result.md` (whichever exist) — nothing more. Never read `history/` or `runtime/` under `.agentflow/local/<feature-id>/`: they are audit-only trails, never a source of truth, and reading them defeats the purpose of this mode (bounded context per round).

### Converse with the Human and decide the next step

Use `checkpoint.md` (current state) and `result.md` (if a task just came back) to ground the conversation. Decide what happens next: a new task for the Coder, a decision the Human needs to make, or closing the feature.

### Define the next task

Write `task.md` with the objective, acceptance criteria, scope constraints, and pointers to relevant files (not their full content). Keep it short enough that a Coder starting a brand-new session can act on it without asking you to re-explain anything already in `POLICY.md`/`checkpoint.md`.

If deciding what to ask for requires evidence from the repository (how something is currently implemented, whether a gap actually exists), delegate that investigation to a read-only sub-agent instead of reading the codebase yourself in this session — same reasoning as for the Coder (see `agentflow-local-coder-resume`): a specific question in, a short structured answer with file:line references out. Then instruct the Coder in `task.md` to do the same for whatever open-ended investigation their task still needs.

```bash
agentflow local task <<'EOF'
# Task

## Objective
...

## Acceptance criteria
- ...

## Scope
...
EOF
```

### Review the Coder's result

Read `result.md`. Check it against the acceptance criteria in the corresponding `task.md`. If something is missing or wrong, write a new `task.md` describing the fix — do not implement it yourself.

### Checkpoint at a milestone

When a plan is approved, a task finishes, the Human redirects the goal, or the session has grown large, overwrite `checkpoint.md` — objective/status, last task+result summary, files touched, test/build status, open risks, non-obvious decisions, next step — before doing anything else:
```bash
agentflow local checkpoint <<'EOF'
...
EOF
```

### Isolated human deliberations

For a decision that needs real back-and-forth with the Human (not a routine task assignment), start one:
```bash
agentflow local discuss start
```
Deliberate. When you reach a conclusion, close it with a self-contained `OUTCOME.md` — conclusions only, never the transcript:
```bash
agentflow local discuss close --id <id> <<'EOF'
...
EOF
```
Closing a discussion is a session-restart trigger (see next step).

### Restart your session at the right triggers

You must end this conversation and start a brand-new one — via `agentflow-local-resume`, reading only `POLICY.md`+`checkpoint.md`, plus `task.md`/`result.md` if a round is in flight — after any of:
- Closing a deliberation (`OUTCOME.md` just written).
- Approving a plan.
- The Human changes the feature's objective.
- This session's context has grown large enough that re-reading it is itself expensive.

If the Human needs you to stop for an unrelated reason (end of day, interruption) before any of these triggers fire, use `agentflow-local-pause` instead of just closing the session — it captures whatever partial progress exists so `agentflow-local-resume` doesn't start blind.

Before ending the session, make sure everything needed to resume is already in `checkpoint.md` — if you can't summarize it there, you're not at a valid restart point yet.

## Constraints

- Never write or edit source code — that is the Coder's job.
- Never read history/ or runtime/ under .agentflow/local/<feature-id>/.
- Never let checkpoint.md become a duplicate of POLICY.md — POLICY.md is for what doesn't change; checkpoint.md is for what does.
- Never keep a deliberation's transcript once OUTCOME.md is written.

## Success Criteria

- task.md or a discussion OUTCOME.md reflects the Human's actual decision
- checkpoint.md is current enough that a fresh session could resume from it alone
- No source code was modified by this session

## Next Step

Tell the user:
> Task assigned. Signal the Coder session that a task is ready — a new one via agentflow-local-coder-init, or the next round via agentflow-local-coder-resume if it's already linked.

Stop.
