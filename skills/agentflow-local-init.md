# AgentFlow Local Init (Controller bootstrap)

You are bootstrapping AgentFlow's local mode for a new feature, as the Controller. This runs once per feature — every session after this one uses agentflow-local-resume instead. It has two stages: Stage A happens in the root repo checkout, before any worktree exists, and ends with the Human approving a PLAN.md. Stage B only starts after that approval.

## Idempotency

If `.agentflow/local/<feature-id>/` already exists inside a worktree for this feature, do not re-run this skill — use agentflow-local-resume instead. If a *pending* `.agentflow/local/<feature-id>/` already exists in the root repo (Stage A was started in an earlier session), resume Stage A from wherever `PLAN.md`/`POLICY.md` left off instead of starting over.

## Steps

### 1. Confirm scope with the Human

By the time this skill runs, the Human should already have decided, in conversation, that a feature will start and roughly what it's for. If that conversation hasn't happened yet, have it first — this skill is not a substitute for deciding scope.

**Fast track exception:** if the Human explicitly asks for it, the fix is small and well understood (a handful of files, no schema/API/contract changes, nothing that needs analysis to even understand), and getting it wrong is low-risk and easily reversible, skip Stage A entirely and go straight to Stage B with a single task — no `PLAN.md`. You never decide to fast-track on your own; the Human must ask for it. If it turns out mid-work to be bigger than expected, stop and ask the Human to switch to the full flow instead of improvising.

### 2. Bootstrap the pending coordination folder

Run this from the **root repo checkout** — there is no worktree yet:
```bash
agentflow local init --repo <root-repo-path> --feature <name> --root <root-branch>
```
This creates `.agentflow/local/<name>/` (POLICY.md skeleton, empty checkpoint.md, required subfolders) right there in the root checkout, and records `<name>` as the root repo's current feature. It moves into the worktree at the end of Stage B — nothing about this location is permanent.

### 3. Draft POLICY.md's stable questions

Open `.agentflow/local/<name>/POLICY.md` and fill in the placeholders directly — it's hand-authored, not piped through a CLI command like checkpoint.md/task.md/result.md/PLAN.md are. At minimum:
- **Validation gate**: check this repo's `AGENTS.md`/`CONTRIBUTING.md` for mandatory test/build commands or a CI/promotion gate; ask the Human if none is documented.
- **Commit conventions**: this feature's conventions beyond the repo's defaults, if any (the generated skeleton already has a fixed "Commit discipline" section — this is about anything additional).

If evidence from the repo is needed to answer either question, delegate that lookup to a read-only sub-agent instead of exploring the codebase yourself in this session. None of this depends on the worktree existing, so it happens now.

### 4. Analyze the problem or need

Talk it through with the Human: what problem or need is this solving, and why now? Capture this as PLAN.md's first section — a short problem statement, not a transcript.

### 5. Assess impact

Delegate to a read-only sub-agent: which files/modules are affected, a rough size-of-effort estimate, and any risk areas (schema/API/contract changes, code with poor test coverage, anything touching shared infrastructure). Give it a specific question and have it report back a short, structured answer with `file:line` references — never full file dumps or an open-ended "explore the codebase" instruction. Do not do this exploration yourself in this session.

### 6. Draft PLAN.md

Write PLAN.md with three sections: Problem/Need (from the analysis step), Impact assessment (from the sub-agent), and a Work plan — a sequenced checklist, one line per step, GFM task-list syntax so progress can be tracked automatically:
```bash
agentflow local plan <<'EOF'
# Plan

## Problem/Need
...

## Impact assessment
- Affected files/modules: ...
- Size of effort: ...
- Risks: ...

## Work plan
- [ ] <first step, short objective>
- [ ] <second step>
...
EOF
```
Each checklist item should be independently completable and independently committable — this is what makes Stage B's task-by-task iteration and partial commits possible.

### 7. Validate the plan from an implementer's perspective

Delegate to a sub-agent explicitly briefed to review PLAN.md the way the Coder would when handed each step cold: Is the sequencing right? Is any step's acceptance criteria vague enough that an implementer would have to guess? Does the impact assessment actually match what the plan proposes to touch? Is anything missing?

If the review surfaces real gaps, revise and re-run `agentflow local plan` (the previous version is archived to `history/` automatically) before moving on — do not present an unreviewed plan to the Human.

### 8. Get the Human's go/no-go

Present PLAN.md to the Human.
- **Approved** → move to Stage B below.
- **Changes requested** → revise PLAN.md (repeat the draft/validate steps above) and present again.
- **No-go** → stop here. Nothing to clean up — the pending folder stays harmlessly in the root repo's gitignored `.agentflow/local/` until the Human revisits it or it's removed later.

### 9. Create the worktree

If the feature's worktree/branch don't exist yet, create them:
```bash
git worktree add .worktrees/<name> -b feature/<name> <root-branch>
cd .worktrees/<name>
```
If the Human already created them, just `cd` into the worktree and verify with `git worktree list`. All following commands run from inside that worktree.

### 10. Promote the pending coordination folder

Move Stage A's folder (POLICY.md, checkpoint.md, PLAN.md, history/) from the root repo into this worktree:
```bash
agentflow local promote --feature <name> --from <root-repo-path>
```
(run from inside the worktree — `--repo` defaults to `.`). This also records `<name>` as this worktree's current feature, so every later `agentflow local` command here can omit `--feature`.

### 11. Fill in POLICY.md's branch/worktree/root fields

These were unknown during Stage A (no worktree existed yet). Edit `.agentflow/local/<name>/POLICY.md` by hand now that `promote`'s output has reminded you they're known: `root_branch`, `branch`, `worktree`.

### 12. Seed the first task from PLAN.md

Write task.md from PLAN.md's **first** unchecked checklist item — objective, acceptance criteria, scope constraints, and pointers to the specific files the impact assessment surfaced (not their full content), so the Coder isn't repeating that exploration from scratch.
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

### 13. Tell the Human to link the Coder

The Human should now open a second, independent session in the same worktree and invoke `agentflow-local-coder-init` there. Give them the exact instruction to do so before continuing — this is the point at which Controller and Coder become linked and ready to work.

### 14. Review the Coder's result

Read `result.md`. Check it against the acceptance criteria in the corresponding `task.md`. If something is missing or wrong, write a new `task.md` describing the fix — do not implement it yourself.

If approved and this feature has a `PLAN.md`, check off the item this task completed (re-run `agentflow local plan` with that line changed from `- [ ]` to `- [x]` — the previous version is archived to `history/` automatically, same as any other round file). If that was the **last** unchecked item, this is full-plan completion, not just a single-task approval: do one final pass confirming every acceptance criterion across the whole plan is actually met, then proceed straight to final review → merge → `agentflow close`, the same way GitHub mode's Phase 3 approval is also the approval to merge.

### 15. Checkpoint at a milestone

When a plan is approved, a task finishes, the Human redirects the goal, or the session has grown large, overwrite `checkpoint.md` — objective/status, last task+result summary, files touched, test/build status, open risks, non-obvious decisions, next step — before doing anything else:
```bash
agentflow local checkpoint <<'EOF'
...
EOF
```

### 16. Isolated human deliberations

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

### 17. Restart your session at the right triggers

You must end this conversation and start a brand-new one — via `agentflow-local-resume`, reading only `POLICY.md`+`checkpoint.md`, plus `task.md`/`result.md` if a round is in flight — after any of:
- Closing a deliberation (`OUTCOME.md` just written).
- Approving a plan.
- The Human changes the feature's objective.
- This session's context has grown large enough that re-reading it is itself expensive.

If the Human needs you to stop for an unrelated reason (end of day, interruption) before any of these triggers fire, use `agentflow-local-pause` instead of just closing the session — it captures whatever partial progress exists so `agentflow-local-resume` doesn't start blind.

Before ending the session, make sure everything needed to resume is already in `checkpoint.md` — if you can't summarize it there, you're not at a valid restart point yet.

In Claude Code this means literally closing this chat and starting a new one with `/agentflow-local-resume` — there is no in-session "soft reset" that achieves the same bounded-context effect.

### 18. Tell the Human

> Bootstrap complete and first task assigned. Tell the Human to open a second session and run agentflow-local-coder-init there.

### 19. Stop

Do not implement the task yourself. Wait for the Human to signal that the Coder has written result.md.
