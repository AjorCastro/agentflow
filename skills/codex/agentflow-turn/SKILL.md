---
name: agentflow-turn
description: Executes the CLI Agent's assigned turn in an AgentFlow workspace. Use when the user says it is the CLI Agent's turn, asks to continue working on a feature, or mentions that the Web Reviewer has approved something. Trigger phrases include "it's your turn", "the web reviewer approved", "continue the feature", "run your turn", "agentflow turn", "do the next step".
---

# AgentFlow Turn

## Core Workflow

1. Pull latest changes from origin
2. Orient yourself
3. Read the exchange folder
4. Read phase context
5. Do the work
6. If you find ambiguity or risk
7. Write a result file
8. Update STATUS.md and state.json
9. Write a handoff note
10. Commit and push

## Idempotency

Before executing any step, check what already exists in `discovery/`, `plans/`, `tasks/`, and `decisions/` that correspond to the current phase. If results already exist for the current phase, do not overwrite them — report what you find and ask the Human whether to continue or start a new task.

`plans/PLAN.md` is the one exception to "do not overwrite": there is always exactly one plan per feature. If it already exists and you're asked to plan again — because scope evolved, not because you're starting a new feature — edit it in place and add an entry to its `## Changelog` section describing what changed and why. Never create a second plan file (e.g. `plan-2.md`, `plan-v2.md`) next to it. If the new work is not a revision of the current plan's scope but something separable, stop and tell the Human this looks like a new feature rather than a plan revision.

## Step Detail

### Pull latest changes from origin

```bash
git pull --rebase
```

### Orient yourself

Read the latest file in `handoffs/` if it exists. Do this every time — it confirms context in ongoing sessions and re-orients you in new ones.

### Read the exchange folder

Find: `.agentflow/features/*/`

Read in order:
1. `CONFIG.md` — feature scope, constraints, policies
2. `STATUS.md` — current phase, turn, next action
3. `state.json` — confirm `current_turn` is `cli`

If `current_turn` is not `cli`, stop and tell the Human:
> "The current turn is not assigned to the CLI Agent. STATUS.md says: [current_turn value]."

If there are files in `decisions/` that have not been resolved, stop and tell the Human:
> "There is an unresolved decision request. Please resolve it before asking me to continue."

### Read phase context

Read everything in the exchange folder that is relevant to the current phase:
- `specs/` — feature specifications
- `discovery/` — research and codebase analysis
- `plans/PLAN.md` — the current implementation plan (single file; check its `## Changelog` for revisions)
- `tasks/` — previous task results
- `reviews/` — feedback from the Web Reviewer

### Do the work

Act according to what `STATUS.md` says is the `next_action`. Use your judgment — the context in the exchange folder is enough to know what is expected.

Core policies (from CONFIG.md — always apply):
- Do not implement anything before discovery and planning have been approved by the Web Reviewer.
- When you find ambiguity or risk, do not guess. Create a decision request instead (see next step).
- There is exactly one plan file: `plans/PLAN.md`. Create it once; on every later revision, edit it in place and add an entry under its `## Changelog` heading instead of creating a new file.
- Once Phase 3 (implementation) is approved — `STATUS.md` says phase `done`, next action mentions merge — that approval already covers the PR and the merge. Open the PR if repo policy requires one, merge it, then run `agentflow close`. Do not create a decision request or otherwise ask the Web Reviewer to re-approve the PR or the merged result; only escalate if something unexpected happens (conflicts, failing CI).
- Write a result file at the end of every task.
- Update STATUS.md and state.json at every handoff.

### If you find ambiguity or risk

Create a file at:
```
decisions/decision-<YYYY-MM-DD>-<short-slug>.md
```

With this content:
```markdown
# Decision Request — <short title>

## Context
<what you were doing when you found the issue>

## Question
<the specific question that needs resolution>

## Options
- Option A: ...
- Option B: ...

## Recommendation
<your recommendation if you have one>
```

Then update STATUS.md and state.json with `current_turn: web` and stop. The Web Reviewer will read the decision file, discuss with the Human if needed, and return the turn to `cli`.

### Write a result file

At the end of every task, write a result file at:
```
tasks/task-<YYYY-MM-DD>-<short-slug>.md
```
Summarizing what you did, what files you created or modified, and any open questions.

### Update STATUS.md and state.json

Update `STATUS.md`:
```markdown
## Current phase
<current or next phase>

## Current turn
web

## Status
<short description of what was completed>

## Next action
<what the Web Reviewer should do>

## Last update
<timestamp>
```

Update `state.json` with the same values (`current_phase`/`current_turn` must match `STATUS.md` literally; `status`/`next_action` are a machine-slug equivalent of the same fact, not a literal copy — see `docs/WEB-AGENT-ROLE.md` for the exact rule).

### Write a handoff note

Write a handoff note at:
```
handoffs/handoff-<YYYY-MM-DD>-codex-cli.md
```
With this content:
```markdown
# Handoff — <phase> — <date>

## Agent
Codex CLI

## What was done
<summary of this turn>

## Files created or modified
<list>

## Next step for the CLI Agent
<exact next action when the turn returns to cli>

## Open questions
<if any, otherwise "none">
```

### Commit and push

```bash
git add -A
git commit -m "cli: <short description of what was done>"
git push
```

## Constraints

- Do not start the next phase after finishing.
- Do not implement before planning is approved.
- One turn = one phase step. Stop when the handoff is done.

## Success Criteria

- Result file exists in `tasks/`
- STATUS.md shows `current_turn: web`
- Changes committed and pushed

## Next Step

Tell the user:
> Done. The Web Reviewer can now review [what you did] on branch [branch name].

Stop.
