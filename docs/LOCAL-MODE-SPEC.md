# AgentFlow — Local Mode Specification (draft)

**Status:** SUPERSEDED by `.agentflow-local/discussions/001/OUTCOME.md` and
`.agentflow-local/discussions/001/PROTOTYPE-PLAN.md`. Kept for historical
context only — do not implement from this document. In particular, §5 of
this draft ("the CLI Agent maintains conversational/session continuity
across tasks") was explicitly rejected after review: the Coder must start a
**clean session per task**, loading only `checkpoint.md` + `task.md` +
`POLICY.md` (see `OUTCOME.md` §3, §5, decision 10 in §11, and §12).
**Audience:** historical — not implementation guidance.

## 1. Problem

The current protocol coordinates the Web role and the CLI Agent through a
shared feature branch on GitHub: Web edits Markdown files via the GitHub web
UI, the CLI Agent pulls/commits/pushes to the same branch. GitHub acts as a
shared filesystem, not as a messaging channel (verified: no `gh pr comment`
or `gh issue` usage anywhere in the protocol — see `internal/protocol/templates.go`,
`internal/skillgen/skills_turn.go`).

Each turn, the CLI Agent re-reads `STATUS.md`, `state.json`, `CONFIG.md`, the
latest handoff, and — depending on phase — `specs/`, `discovery/`,
`plans/PLAN.md`, and **all accumulated** `tasks/*.md` and `reviews/*.md`.
The last two grow without bound or pruning across the life of a feature and
are the main suspected driver of rising input-token cost per turn.

## 2. Goal

Define an alternative coordination mode ("local mode") between a
**Controller** agent (talks to the human, directs work, assigns tasks,
reviews results) and the existing **CLI Agent** role (investigates, plans,
implements, tests) that minimizes input tokens re-sent to the LLMs on each
round, without using GitHub as the communication channel.

## 3. Scope

Local mode is an **alternative, opt-in mode**, coexisting with the current
GitHub-based protocol. It must not change existing behavior when not
selected. Exact selection mechanism (flag / config field) is left to the
implementer; a natural candidate is a mode field on `agentflow init` (e.g.
`--mode local`, default remains the current GitHub-based flow).

The CLI Agent's responsibilities and skills (investigate, plan, implement,
test) are unchanged. What changes is only the coordination channel and the
lifecycle of the artifacts that carry state between Controller and CLI
Agent.

## 4. Process model

- Controller and CLI Agent run as **two independent, long-running local
  processes** (not one spawning the other as a subprocess it blocks on).
- Turn-taking (deciding when the Controller hands off to the CLI Agent, and
  when the CLI Agent hands a result back to the Controller) is **signaled
  manually by the human** in v1. No file-watching/daemon-polling mechanism
  is required for this iteration.
- GitHub is not part of the coordination channel in this mode. `git` is
  still used normally for the actual code (branch, commits, eventual merge
  to `develop`) — only the *coordination* artifacts move out of Git.

## 5. Core mechanism for reducing token cost

The CLI Agent, being long-running, **maintains conversational/session
continuity across tasks** instead of being re-invoked stateless per task.
It does not need to re-read full accumulated history on every round — only
the new instruction from the Controller.

To keep the CLI Agent's own context from growing unbounded over a long
feature, it periodically writes a **checkpoint**: a compact, self-contained
summary that fully replaces the need to replay prior conversation turns.

- **Checkpoint trigger:** natural milestones in the work (e.g., a task is
  completed, a plan is approved, before a risky change, before a handoff) —
  decided by the Controller, **not** a fixed turn-count or size cadence.
- **Checkpoint content (minimum):**
  - Feature goal and overall status (done / pending)
  - Last task assigned and its result (summarized, not full detail)
  - Files touched so far (list, not full diff)
  - Test/build status (pass / fail / not run, last relevant command)
  - Open blockers or risks (anything left mid-way, unresolved errors,
    pending decisions)
  - Non-obvious design decisions made so far
  - Suggested next step
- The checkpoint must be sufficient, on its own (plus the latest task/result
  exchange, see §6), for **either** a restarted CLI Agent session **or** a
  fresh Controller session to resume work without the human re-explaining
  context from scratch. This is a design requirement to validate during
  implementation, not something guaranteed by construction — if it proves
  insufficient in practice, the checkpoint content should be extended.

## 6. Artifacts and layout

All local-mode coordination artifacts:

- Live **inside the feature worktree**, e.g. under
  `.agentflow/local/<feature-id>/` (exact path/naming left to implementer,
  consistent with the existing `.agentflow/features/<feature-id>/`
  convention for the GitHub mode).
- Are **excluded from Git** (`.gitignore` entry) — they are coordination
  scaffolding, not project history.
- Are **ephemeral**: deleted together with the worktree/branch on
  `agentflow close`, exactly like today. The only permanent record of the
  work remains the code's commit history on `develop` — this mode does not
  change that.

Proposed file set (draft, adjust freely during implementation):

| File | Written by | Overwritten or appended | Re-read automatically? |
|---|---|---|---|
| `task.md` | Controller | overwritten each round | yes, by CLI Agent |
| `result.md` | CLI Agent | overwritten each round | yes, by Controller |
| `checkpoint.md` | CLI Agent | overwritten at milestones | yes, by both, for recovery |
| `history/` | both | append-only, timestamped files | **no** — audit/debug trail for humans only, never fed back into LLM context automatically |

No per-phase folders (`specs/`, `discovery/`, `plans/`, `decisions/`,
`reviews/`) — local mode is phase-fluid. The Controller manages phase
progression conversationally with the human; `checkpoint.md` is the single
source of truth for "where are we", not a set of per-phase documents.

## 7. Recovery behavior

- If the CLI Agent's session is lost/restarted, it rehydrates from
  `checkpoint.md` (plus `task.md` if a task was in flight).
- If the Controller's session is lost/restarted, it rehydrates from
  `checkpoint.md` + `task.md` + `result.md`. No separate Controller-specific
  checkpoint is planned for v1 — validate during implementation whether this
  is actually sufficient; if not, add a minimal Controller-side checkpoint.

## 8. Explicitly open / deferred (not decided in this deliberation)

- Exact CLI subcommands/flags to enable and configure local mode.
- Exact file/folder naming (the table in §6 is a starting proposal).
- How Controller and CLI Agent processes are concretely launched (two
  Claude Code sessions? wrapped by `agentflow` subcommands?).
- Whether/how `CONFIG.md`-style fixed policy text still applies in this
  mode, or is replaced by something else.
- Whether a Controller-specific checkpoint turns out to be necessary in
  practice (see §7).
- Any tooling to help the human know when it's the right moment to hand off
  a turn (still manual in v1, per §4).

## 9. Facts vs. hypotheses used in this design

**Verified in code** (via investigation of this repo): no GitHub
messaging (`gh pr comment`/`gh issue`) exists in the current protocol;
GitHub is used only as a shared filesystem via branch commits; `tasks/*.md`
and `reviews/*.md` accumulate unpruned and are read in full each turn;
`README.md` claims "no GitHub integration" which contradicts the rest of
the protocol (likely stale documentation, not a real mode).

**Hypothesis** (not verified): that `tasks/`/`reviews/` accumulation is the
main driver of token cost growth, rather than e.g. repeated full-file reads
of `plans/PLAN.md` or general session overhead. Worth confirming with
before/after token measurements once local mode is implemented.
