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
| `agentflow-local-init` | Controller | Once, right after the Human decides in conversation that a feature will start. Two stages: Stage A (pre-worktree) analyzes the problem/need, assesses impact, and gets a `PLAN.md` approved by the Human; Stage B (post-approval) creates the worktree, promotes `.agentflow/local/<id>/` into it, drafts `POLICY.md`, and seeds the first task. A Human-requested fast track skips Stage A for small, well-understood fixes. |
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
excluded from Git. During Stage A this folder is *pending*, in the root repo
checkout (no worktree exists yet); `agentflow local promote` moves it into
the feature worktree at the start of Stage B. `agentflow close` deletes the
whole folder from the worktree — nothing to clean up by hand.

| File | Written by | Overwritten or appended | Read automatically by |
|---|---|---|---|
| `POLICY.md` | Controller, once at `local init` (Stage A) | overwritten only if a policy changes | both, on every session start |
| `PLAN.md` | Controller, during Stage A, via `agentflow local plan` | overwritten at revisions and when checking off items | Controller only — kept out of the Coder's context on purpose |
| `checkpoint.md` | Coder (and Controller when relevant) | overwritten at milestones | both, on every session start |
| `task.md` | Controller | overwritten each round | Coder |
| `result.md` | Coder | overwritten each round | Controller |
| `discussions/<id>/OUTCOME.md` | Controller, via `discuss close` | written once, never edited after | Controller, next session after closing that deliberation |
| `history/` | both, automatically | append-only, timestamped copies of prior task/result/checkpoint/plan | nobody — audit trail for the Human only |

`POLICY.md` holds what doesn't change turn to turn: feature ID, branch,
worktree, root, validation gate, commit conventions (including a fixed
"Commit discipline" rule — commit per completed `PLAN.md` item, never
batched), scope restrictions. `PLAN.md` holds the problem statement, impact
assessment, and sequenced checklist agreed with the Human before any code
was written — it's what keeps the work sequence intact across Controller/
Coder session restarts, instead of relying solely on `checkpoint.md` (which
is overwritten, not accumulated, at every milestone). `checkpoint.md` holds
what changes turn to turn: overall status, last task/result summary, files
touched, test/build state, open risks, next suggested step.

## Quick start

1. Decide the feature with the Human. In a Controller session, invoke the
   `agentflow-local-init` skill — it runs in two stages.

   **Stage A** (from the root repo checkout, before any worktree exists):
   bootstraps a pending coordination folder, analyzes the problem/need with
   the Human, delegates an impact assessment to a sub-agent, drafts and
   validates `PLAN.md`, and gets the Human's go/no-go:
   ```
   agentflow local init --repo <root-repo-path> --feature <name>
   agentflow local plan <<'EOF'
   ...
   EOF
   ```
   A Human-requested fast track (small, well-understood, low-risk fix) skips
   Stage A entirely — straight to Stage B with a single task and no `PLAN.md`.

   **Stage B** (only after Stage A is approved): creates the worktree,
   promotes the pending folder into it, fills in `POLICY.md`'s
   branch/worktree/root fields plus validation gate and commit conventions,
   and seeds the first `task.md` from `PLAN.md`'s first checklist item:
   ```
   agentflow local promote --feature <name> --from <root-repo-path>
   agentflow local task <<'EOF'
   ...
   EOF
   ```

2. In a second, independent session opened in the same worktree, invoke the
   `agentflow-local-coder-init` skill. It confirms the link and loads
   exactly `POLICY.md` + `checkpoint.md` + `task.md` via:
   ```
   agentflow local context --role coder
   ```
   investigates, plans, implements, tests, commits (one commit for this
   task/`PLAN.md` item, per `POLICY.md`'s commit discipline), then reports
   back:
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
   check off the completed item in `PLAN.md`, and either assign the next
   task (`PLAN.md`'s next unchecked item) or close the feature. Check
   progress any time with:
   ```
   agentflow local status
   ```
   which shows `PLAN.md`'s checklist progress as `N/M done` when a plan
   exists.

4. From here on, every session uses the **resume** skills, not the **init**
   ones: `agentflow-local-resume` for the Controller, `agentflow-local-coder-resume`
   for the Coder's next task. The Controller restarts its session (fresh
   `agentflow-local-resume` invocation) after closing a deliberation,
   approving a plan, or when the Human redirects the goal — recovering
   purely from `POLICY.md`+`PLAN.md`+`checkpoint.md`+`task.md`+`result.md`,
   never from replaying the conversation.

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

7. When every `PLAN.md` item is checked off (or the fast-tracked task is
   done), close the feature exactly like GitHub mode:
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
