---
name: agentflow-turn
description: Executes the CLI Agent's turn in an AgentFlow workspace. Use this whenever the user says it is the CLI Agent's turn to act, or asks you to continue working on a feature.
---

## Objective

Read the current state of the AgentFlow exchange folder and act according to what is expected in the current phase. Commit and push when done.

## Idempotency

Before executing any step, check what already exists in `discovery/`, `plans/`, `tasks/`, and `decisions/`. If results already exist for the current phase, do not overwrite them — report what you found and ask the user whether to continue or start a new task.

## Procedure

### 1. Read the exchange folder

Find the exchange folder: `.agentflow/features/*/`

Read in order:
1. `CONFIG.md` — feature scope, constraints, policies
2. `STATUS.md` — current phase, turn, next action
3. `state.json` — confirm `current_turn` is `cli`

If `current_turn` is not `cli`, stop and tell the user:
> "The current turn is not assigned to the CLI Agent. STATUS.md says: [value]."

If there are unresolved files in `decisions/`, stop and tell the user:
> "There is an unresolved decision request. Please resolve it before asking me to continue."

### 2. Read all relevant context

Depending on the current phase, read:
- `specs/` — requirements
- `discovery/` — research and codebase analysis
- `plans/` — implementation plans
- `tasks/` — previous results
- `reviews/` — Web Reviewer feedback

### 3. Do the work

Act according to `STATUS.md` `next_action`. Use your judgment — the exchange folder provides enough context to know what is expected.

**Core policies (always apply):**
- Do not implement anything before discovery and planning are approved by the Web Reviewer.
- When you find ambiguity or risk, do not guess — create a decision request (see below).
- Write a result file at the end of every task.
- Update STATUS.md and state.json at every handoff.

### 4. If you find ambiguity or risk

Create a file at `decisions/decision-<YYYY-MM-DD>-<short-slug>.md`:

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

### 5. When you finish

1. Write a result file at `tasks/task-<YYYY-MM-DD>-<short-slug>.md` summarizing what you did.

2. Update `STATUS.md`:
   - `current_turn: web`
   - `next_action`: what the Web Reviewer should do
   - `last_update`: current timestamp

3. Update `state.json` with the same values.

4. Commit and push:
   ```
   git add -A
   git commit -m "cli: <short description>"
   git push
   ```

## Success criteria

- Result file exists in `tasks/`
- `STATUS.md` shows `current_turn: web`
- Changes committed and pushed to GitHub

## Next step

Tell the user:
> "Done. The Web Reviewer can now review [what you did] on branch [branch name]."

Stop. Do not continue to the next phase.
