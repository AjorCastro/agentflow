---
name: agentflow-turn
description: Executes the CLI Agent's assigned turn in an AgentFlow workspace. Use when the user says it is the CLI Agent's turn, asks to continue working on a feature, or mentions that the Web Reviewer has approved something. Trigger phrases include "it's your turn", "the web reviewer approved", "continue the feature", "run your turn", "agentflow turn", "do the next step".
---

# AgentFlow Turn

## Core Workflow

1. Read the exchange folder to understand the current state.
2. Confirm `current_turn` is `cli`.
3. Check for unresolved decision requests.
4. Read all relevant context for the current phase.
5. Do the work. Create a result file.
6. Update STATUS.md and state.json.
7. Commit and push.

## Idempotency

Before acting, check what already exists in `discovery/`, `plans/`, `tasks/`, and `decisions/`. If results already exist for the current phase, do not overwrite them. Report what you found and ask the user whether to continue or start a new task.

## Step Detail

### Read exchange folder

Find: `.agentflow/features/*/`

Read in order:
1. `CONFIG.md` — feature scope, constraints, policies
2. `STATUS.md` — current phase, turn, next action
3. `state.json` — confirm `current_turn` is `cli`

If `current_turn` is not `cli`:
> Stop: "The current turn is not assigned to the CLI Agent. STATUS.md says: [value]."

If unresolved files exist in `decisions/`:
> Stop: "There is an unresolved decision request. Please resolve it before asking me to continue."

### Read phase context

- `specs/` — requirements and scope
- `discovery/` — codebase analysis
- `plans/` — approved implementation plan
- `tasks/` — previous results
- `reviews/` — Web Reviewer feedback

### Do the work

Act according to `STATUS.md` `next_action`. The exchange folder provides enough context to know what is expected.

**Core policies — always apply:**
- Do not implement anything before discovery and planning are approved by the Web Reviewer.
- When you find ambiguity or risk, create a decision request instead of guessing.
- Write a result file at the end of every task.
- Update STATUS.md and state.json at every handoff.

### Decision requests

When ambiguity or risk is found, create `decisions/decision-<YYYY-MM-DD>-<slug>.md`:

```markdown
# Decision Request — <short title>

## Context
<what you were doing>

## Question
<the specific question>

## Options
- Option A: ...
- Option B: ...

## Recommendation
<your recommendation if any>
```

Update STATUS.md with `current_turn: web` and stop. The Web Reviewer will read the decision file, discuss with the Human if needed, and return the turn to `cli`.

### Write result file

Create `tasks/task-<YYYY-MM-DD>-<slug>.md` summarizing:
- What was done
- Files created or modified
- Open questions if any

### Update STATUS.md and state.json

```
current_turn: web
next_action: <what the Web Reviewer should do>
last_update: <timestamp>
```

### Commit and push

```
git add -A
git commit -m "cli: <short description>"
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
> "Done. The Web Reviewer can now review [what you did] on branch [branch]."

Stop.
