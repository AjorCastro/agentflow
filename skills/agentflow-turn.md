# AgentFlow Turn

It is your turn to act. Read the current state of the exchange folder and do what is expected of you.

## Idempotency

Before executing any step, check if it was already done. Look for existing files in `discovery/`, `plans/`, `tasks/`, and `decisions/` that correspond to the current phase. If results already exist for the current phase, do not overwrite them — report what you find and ask the Human whether to continue or start a new task.

## Before you start

Read these files in order:

1. `.agentflow/features/*/CONFIG.md` — understand the feature, scope, and policies
2. `.agentflow/features/*/STATUS.md` — understand current phase, turn, and next action
3. `.agentflow/features/*/state.json` — confirm `current_turn` is `cli`

If `current_turn` is not `cli`, stop and tell the Human:
> "The current turn is not assigned to the CLI Agent. STATUS.md says: [current_turn value]."

If there are files in `decisions/` that have not been resolved, stop and tell the Human:
> "There is an unresolved decision request. Please resolve it before asking me to continue."

## Do the work

Read everything in the exchange folder that is relevant to the current phase:
- `discovery/` — research and codebase analysis
- `plans/` — implementation plans
- `tasks/` — previous task results
- `reviews/` — feedback from the Web Reviewer
- `specs/` — feature specifications

Act according to what `STATUS.md` says is the `next_action`. Use your judgment — the context in the exchange folder is enough to know what is expected.

### Core policies (from CONFIG.md — always apply)

- Do not implement anything before discovery and planning have been approved by the Web Reviewer.
- When you find ambiguity or risk, do not guess. Create a decision request instead (see below).
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

## When you finish

1. Write a result file at:
   ```
   tasks/task-<YYYY-MM-DD>-<short-slug>.md
   ```
   Summarizing what you did, what files you created or modified, and any open questions.

2. Update `STATUS.md`:
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

3. Update `state.json` with the same values.

4. Commit and push:
   ```bash
   git add -A
   git commit -m "cli: <short description of what was done>"
   git push
   ```

5. Tell the Human:
   > "Done. The Web Reviewer can now review [what you did] on branch [branch name]."

## Stop

Do not start the next phase or continue working. Your turn is complete.
