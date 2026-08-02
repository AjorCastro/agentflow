# AgentFlow — Local Mode Guide

Local mode is an **alternative** to the GitHub-based protocol (see
`docs/WEB-AGENT-ROLE.md`), for when there is no Web Reviewer and both
coordinating roles run locally. It coordinates two agent sessions —
**Controller** and **Coder** — through a handful of small files inside the
feature worktree instead of a GitHub exchange folder.

Design source: `.agentflow-local/discussions/001/OUTCOME.md` (decisions) and
`PROTOTYPE-PLAN.md` (file formats, exact cycle). This guide is the practical
walkthrough; those documents are authoritative if anything here conflicts.

## Roles

| Role | Where it runs | Responsibility |
|---|---|---|
| Human | Terminal | Directs work, signals turns between sessions |
| **Controller** | One Claude Code (or similar) session | Talks to the Human, assigns tasks, reviews results. **Never writes code.** |
| **Coder** | A second, independent session, in the same worktree | Investigates, plans, implements, tests |

Controller and Coder are separate **sessions**, not one calling the other.
Turn-taking is manual: the Human tells each session when it's their turn.

## Core rule

The Coder starts a **clean session per task** — it never continues a prior
conversation. It rehydrates entirely from `POLICY.md` + `checkpoint.md` +
`task.md`. The Controller restarts its own session after closing a
deliberation, after a plan approval, after the Human redirects the goal, or
whenever its context has grown large — see `OUTCOME.md` §16 for the full
list. In every case, everything needed to resume must already be on disk
before the session ends.

## The files

All coordination artifacts live under `.agentflow/local/<feature-id>/`,
inside the feature worktree, excluded from Git. `agentflow close` deletes
this whole folder — nothing to clean up by hand.

| File | Written by | Overwritten or appended | Read automatically by |
|---|---|---|---|
| `POLICY.md` | Controller, once at `local init` | overwritten only if a policy changes | both, on every session start |
| `checkpoint.md` | Coder (and Controller when relevant) | overwritten at milestones | both, on every session start |
| `task.md` | Controller | overwritten each round | Coder |
| `result.md` | Coder | overwritten each round | Controller |
| `discussions/<id>/OUTCOME.md` | Controller, via `discuss close` | written once, never edited after | Controller, next session after closing that deliberation |
| `history/` | both, automatically | append-only, timestamped copies of prior task/result/checkpoint | nobody — audit trail for the Human only |

`POLICY.md` holds what doesn't change turn to turn: feature ID, branch,
worktree, root, validation gate, commit conventions, scope restrictions.
`checkpoint.md` holds what does: overall status, last task/result summary,
files touched, test/build state, open risks, next suggested step.

## Quick start

1. Create the feature normally:
   ```
   git worktree add .worktrees/<name> -b feature/<name> develop
   cd .worktrees/<name>
   ```

2. Create the coordination folder:
   ```
   agentflow local init --feature <name> --branch feature/<name> --worktree .worktrees/<name>
   ```
   This writes `POLICY.md` and an empty `checkpoint.md`.

3. Edit `.agentflow/local/<name>/POLICY.md` by hand — it's written once, not
   through a command. Fill in the validation gate (equivalent to what
   `AGENTS.md`/`CONTRIBUTING.md` would define for GitHub mode) and any
   commit or scope conventions.

4. In one session, invoke the `agentflow-local-controller` skill. It talks
   to you, decides the first task, and writes it:
   ```
   agentflow local task --feature <name> <<'EOF'
   ...
   EOF
   ```

5. In a second, independent session opened in the same worktree, invoke the
   `agentflow-local-coder` skill. It loads exactly `POLICY.md` +
   `checkpoint.md` + `task.md` via:
   ```
   agentflow local context --feature <name> --role coder
   ```
   investigates, plans, implements, tests, then reports back:
   ```
   agentflow local result --feature <name> <<'EOF'
   ...
   EOF
   ```
   and writes/overwrites `checkpoint.md` (end of task is a mandatory
   checkpoint). This session then ends — the next task is a fresh Coder
   session, never a continuation.

6. Back in the Controller session (or a fresh one if a restart trigger
   fired), review `result.md`, discuss with the Human, and either assign the
   next task (back to step 4) or close the feature. Check progress any time
   with:
   ```
   agentflow local status --feature <name>
   ```

7. If a decision needs real deliberation with the Human, isolate it instead
   of letting it grow the Controller's working context indefinitely:
   ```
   agentflow local discuss start --feature <name>
   # ... deliberate ...
   agentflow local discuss close --feature <name> --id 001 <<'EOF'
   ...
   EOF
   ```
   Closing a discussion is a Controller session-restart trigger — the next
   session recovers from `OUTCOME.md` + `checkpoint.md` alone, not from
   replaying the deliberation.

8. When the feature is merged, close it exactly like GitHub mode:
   ```
   agentflow close --branch feature/<name> --root develop
   ```
   This also removes `.agentflow/local/<name>/` — nothing to clean up by
   hand.

## When to use local mode vs. GitHub mode

Use local mode when there is no Web Reviewer role and both coordinating
agents run on the same machine — it avoids the growing, unpruned
`tasks/*.md`/`reviews/*.md` history that GitHub mode re-reads in full every
turn. Use GitHub mode (the default) when a Web Reviewer needs to
participate via a browser-based AI with GitHub access.

## Status

Local mode is a working prototype (v0), validated once end-to-end on a real
feature (`.agentflow-local/discussions/001/RUN-001.md`). Confirmed:
`checkpoint.md`+`task.md`+`result.md` were sufficient for a fresh Coder
session to act without extra context, and `agentflow close` cleans up the
local folder correctly. Not yet validated: a forced Controller session
restart mid-feature, a real human deliberation closed with `discuss close`,
and the token-cost savings hypothesis that motivated this mode.
