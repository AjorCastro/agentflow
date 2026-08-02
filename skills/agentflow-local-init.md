# AgentFlow Local Init (Controller bootstrap)

You are bootstrapping AgentFlow's local mode for a new feature, as the Controller. This only runs once per feature — every session after this one uses agentflow-local-resume instead.

## Idempotency

If `.agentflow/local/<feature-id>/` already exists for this feature, do not re-run this skill — use agentflow-local-resume instead. This skill is only for a feature that has never been bootstrapped.

## Steps

### 1. Confirm scope with the Human

By the time this skill runs, the Human should already have decided, in conversation, that a feature will start and roughly what it's for. If that conversation hasn't happened yet, have it first — this skill is not a substitute for deciding scope.

### 2. Create or verify the worktree

If the feature's worktree/branch don't exist yet, create them:
```bash
git worktree add .worktrees/<name> -b feature/<name> <root-branch>
cd .worktrees/<name>
```
If the Human already created them, just `cd` into the worktree and verify with `git worktree list`. All following commands in this skill run from inside that worktree.

### 3. Run agentflow local init

```bash
agentflow local init --feature <name> --branch feature/<name> --worktree .worktrees/<name> --root <root-branch>
```
This creates `.agentflow/local/<name>/` (POLICY.md skeleton, empty checkpoint.md, required subfolders) and records `<name>` as this worktree's current feature — every `agentflow local` command run from this worktree afterwards can omit `--feature`.

### 4. Draft POLICY.md with the Human

Open `.agentflow/local/<name>/POLICY.md` and fill in the placeholders directly — it's hand-authored, not piped through a CLI command like checkpoint.md/task.md/result.md are. At minimum:
- **Validation gate**: check this repo's `AGENTS.md`/`CONTRIBUTING.md` for mandatory test/build commands or a CI/promotion gate; ask the Human if none is documented.
- **Commit conventions**: this feature's conventions beyond the repo's defaults, if any.

If evidence from the repo is needed to answer either question, delegate that lookup to a read-only sub-agent instead of exploring the codebase yourself in this session.

### 5. Define the first task

Same as any other task definition: objective, acceptance criteria, scope constraints, pointers to relevant files.
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

### 6. Tell the Human to link the Coder

The Human should now open a second, independent session in the same worktree and invoke `agentflow-local-coder-init` there. Give them the exact instruction to do so before continuing — this is the point at which Controller and Coder become linked and ready to work.

### 7. Review the Coder's result

Read `result.md`. Check it against the acceptance criteria in the corresponding `task.md`. If something is missing or wrong, write a new `task.md` describing the fix — do not implement it yourself.

### 8. Checkpoint at a milestone

When a plan is approved, a task finishes, the Human redirects the goal, or the session has grown large, overwrite `checkpoint.md` — objective/status, last task+result summary, files touched, test/build status, open risks, non-obvious decisions, next step — before doing anything else:
```bash
agentflow local checkpoint <<'EOF'
...
EOF
```

### 9. Isolated human deliberations

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

### 10. Restart your session at the right triggers

You must end this conversation and start a brand-new one — via `agentflow-local-resume`, reading only `POLICY.md`+`checkpoint.md`, plus `task.md`/`result.md` if a round is in flight — after any of:
- Closing a deliberation (`OUTCOME.md` just written).
- Approving a plan.
- The Human changes the feature's objective.
- This session's context has grown large enough that re-reading it is itself expensive.

If the Human needs you to stop for an unrelated reason (end of day, interruption) before any of these triggers fire, use `agentflow-local-pause` instead of just closing the session — it captures whatever partial progress exists so `agentflow-local-resume` doesn't start blind.

Before ending the session, make sure everything needed to resume is already in `checkpoint.md` — if you can't summarize it there, you're not at a valid restart point yet.

In Claude Code this means literally closing this chat and starting a new one with `/agentflow-local-resume` — there is no in-session "soft reset" that achieves the same bounded-context effect.

### 11. Tell the Human

> Bootstrap complete and first task assigned. Tell the Human to open a second session and run agentflow-local-coder-init there.

### 12. Stop

Do not implement the task yourself. Wait for the Human to signal that the Coder has written result.md.
