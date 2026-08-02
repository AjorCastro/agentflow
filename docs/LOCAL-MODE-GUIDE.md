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

## The five skills

Each role has an **init** skill (run once per feature) and a **resume**
skill (run for every session after that, including forced restarts). A
fifth skill, **pause**, is shared by both roles for stopping safely off a
natural milestone.

| Skill | Role | When |
|---|---|---|
| `agentflow-local-init` | Controller | Once, right after the Human decides in conversation that a feature will start. Bootstraps the worktree, `.agentflow/local/<id>/`, `POLICY.md`, and the first task. |
| `agentflow-local-coder-init` | Coder | Once, right after the Controller's bootstrap — confirms the link, then does the first task. |
| `agentflow-local-resume` | Controller | Every session after the first — including after the Controller's own session restarts mid-feature. |
| `agentflow-local-coder-resume` | Coder | Every task after the first — the Coder always starts a clean session per task, by design. |
| `agentflow-local-pause` | Either | When the Human needs to stop a session for a reason unrelated to task/plan completion (end of day, an interruption), so the next resume doesn't start blind. |

## Core rule

The Coder starts a **clean session per task** — it never continues a prior
conversation. It rehydrates entirely from `POLICY.md` + `checkpoint.md` +
`task.md`. The Controller restarts its own session after closing a
deliberation, after a plan approval, after the Human redirects the goal, or
whenever its context has grown large — see `OUTCOME.md` §16 for the full
list. In every case, everything needed to resume must already be on disk
before the session ends.

## Auto-linking: no more copying feature IDs between sessions

`agentflow local init` records the feature ID as this **worktree's current
feature** (`.agentflow/local/CURRENT`). Every other `agentflow local`
subcommand — `task`, `result`, `checkpoint`, `discuss`, `status`,
`context` — falls back to that marker when `--feature` is omitted. Since
Controller and Coder always run in the same worktree, neither session needs
to be told the feature ID by hand; pass `--feature <id>` explicitly only if
you need to target a different feature than the current one.

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

1. Decide the feature with the Human, then bootstrap it: in a Controller
   session, invoke the `agentflow-local-init` skill. It creates the
   worktree if it doesn't exist yet, runs:
   ```
   agentflow local init --feature <name> --branch feature/<name> --worktree .worktrees/<name>
   ```
   drafts `POLICY.md` with you (validation gate, commit conventions), and
   writes the first `task.md`.

2. In a second, independent session opened in the same worktree, invoke the
   `agentflow-local-coder-init` skill. It confirms the link and loads
   exactly `POLICY.md` + `checkpoint.md` + `task.md` via:
   ```
   agentflow local context --role coder
   ```
   investigates, plans, implements, tests, then reports back:
   ```
   agentflow local result <<'EOF'
   ...
   EOF
   ```
   and writes/overwrites `checkpoint.md` (end of task is a mandatory
   checkpoint). This session then ends — the next task is a fresh Coder
   session (`agentflow-local-coder-resume`), never a continuation.

3. Back in the Controller session — or a fresh one via `agentflow-local-resume`
   if a restart trigger fired — review `result.md`, discuss with the Human,
   and either assign the next task or close the feature. Check progress any
   time with:
   ```
   agentflow local status
   ```

4. From here on, every session uses the **resume** skills, not the **init**
   ones: `agentflow-local-resume` for the Controller, `agentflow-local-coder-resume`
   for the Coder's next task. The Controller restarts its session (fresh
   `agentflow-local-resume` invocation) after closing a deliberation,
   approving a plan, or when the Human redirects the goal — recovering
   purely from `POLICY.md`+`checkpoint.md`+`task.md`+`result.md`, never from
   replaying the conversation.

5. If the Human needs to interrupt a session for a reason unrelated to a
   natural milestone (end of day, an unplanned interruption), invoke
   `agentflow-local-pause` first. It writes a partial `checkpoint.md`
   describing in-progress work before the session ends, so the next
   `-resume` doesn't start blind.

6. If a decision needs real deliberation with the Human, isolate it instead
   of letting it grow the Controller's working context indefinitely:
   ```
   agentflow local discuss start
   # ... deliberate ...
   agentflow local discuss close --id 001 <<'EOF'
   ...
   EOF
   ```
   Closing a discussion is a Controller session-restart trigger — the next
   session recovers from `OUTCOME.md` + `checkpoint.md` alone, not from
   replaying the deliberation.

7. When the feature is merged, close it exactly like GitHub mode:
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

Local mode is a working prototype, validated twice end-to-end on real
features:

- `.agentflow-local/discussions/001/RUN-001.md`: confirmed
  `checkpoint.md`+`task.md`+`result.md` are sufficient for a fresh Coder
  session to act without extra context, and that `agentflow close` cleans up
  the local folder correctly.
- `.agentflow-local/discussions/001/RUN-002.md`: confirmed a forced
  Controller session restart mid-feature recovers correctly from
  `POLICY.md`+`checkpoint.md`+`task.md`+`result.md` alone, twice in a row,
  with no human re-explanation needed.

Not yet validated: a real human deliberation closed with `discuss close`,
the `agentflow-local-pause` skill in an actual interrupted session, and the
token-cost savings hypothesis that motivated this mode in the first place.
