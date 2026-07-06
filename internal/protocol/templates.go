package protocol

import (
	"fmt"
	"time"
)

// ReadmeMD returns the content for the exchange README.md.
func ReadmeMD(featureID, title string) string {
	return fmt.Sprintf(`# AgentFlow Exchange — %s

## What is this folder?

This is the **exchange folder** for the AgentFlow protocol. It is the shared workspace
that coordinates work between three roles operating on a single feature branch.

## Roles

| Role         | Responsibility                                                   |
|--------------|------------------------------------------------------------------|
| Human        | Provides requirements, approves decisions, resolves ambiguity    |
| Web Reviewer | Reviews plans and outputs via web UI; drives approval workflow   |
| CLI Agent    | Executes tasks inside the worktree; reads and writes this folder |

## Core rule

> **CLI executes. Web/Human review, decide, and approve.**

The CLI Agent must not implement anything before discovery and planning are reviewed
and approved by the Web Reviewer or Human.

## Folder layout

| Path           | Purpose                                              |
|----------------|------------------------------------------------------|
| STATUS.md      | Current phase, turn, status and next action          |
| CONFIG.md      | Feature metadata and protocol policies               |
| state.json     | Machine-readable state (same info as STATUS.md)      |
| specs/         | Feature specifications and requirements              |
| discovery/     | Research outputs, codebase analysis                  |
| plans/         | PLAN.md — the one current implementation plan, edited in place |
| tasks/         | Task result files written by the CLI Agent           |
| decisions/     | Decision requests and their resolutions              |
| reviews/       | Review outputs from Web Reviewer                     |
| handoffs/      | Handoff documents between turns                      |
| prompts/       | Reusable prompt templates                            |

## Lifecycle

This folder is **temporary to the feature**. It lives on the feature branch and is
merged or discarded after the feature is complete, unless an explicit consolidation
decision is made.

---

feature_id: %s
title: %s
`, title, featureID, title)
}

// StatusMD returns the initial content for STATUS.md.
func StatusMD() string {
	return fmt.Sprintf(`# AgentFlow Status

## Current phase
intake

## Current turn
web

## Status
Initialized.

## Next action
Read the exchange folder and begin Phase 1 — Discovery.

## Last update
%s
`, time.Now().UTC().Format(time.RFC3339))
}

// WebAgentRoleMD returns the content for prompts/web-agent-role.md.
func WebAgentRoleMD() string {
	return `# AgentFlow — Web Agent Role

## Who you are

You are the **Web Reviewer** in the AgentFlow protocol. You work inside a web-based AI
interface (Claude.ai, ChatGPT, Gemini, or similar) that has access to this GitHub repository.

You collaborate with two other roles:

| Role             | Where they work            | Responsibility                                       |
|------------------|----------------------------|------------------------------------------------------|
| Human            | Terminal / IDE             | Provides requirements, approves decisions, runs CLI  |
| **Web Reviewer** | **Web AI + GitHub**        | **Defines work, reviews outputs, approves phases**   |
| CLI Agent        | Worktree (Claude Code etc) | Executes tasks, writes results, commits              |

---

## Core rule

> **CLI executes. You and the Human review, decide, and approve.**

Never ask the CLI Agent to implement anything before discovery and planning are approved.

---

## Where things happen

Everything you read and write lives on the **feature branch** — ` + "`STATUS.md`" + `, ` + "`state.json`" + `, ` + "`discovery/`" + `, ` + "`plans/`" + `, ` + "`tasks/`" + `, ` + "`decisions/`" + `, ` + "`reviews/`" + `. Navigate to that branch on GitHub before reading or editing anything below.

The one exception is Phase 0: ` + "`agentflow-init-<branch-slug>.md`" + ` is created on ` + "`develop`" + ` (the feature branch doesn't exist yet). Everything after that — for the rest of the feature's life — is on the feature branch.

---

## Your tasks by phase

### Phase 0 — Feature definition (before agentflow init)

The Human describes what they want to build. Your job:

1. Ask clarifying questions until you understand scope and constraints.
2. Propose: title, branch name (e.g. ` + "`feature/my-feature`" + `), worktree path (e.g. ` + "`.worktrees/my-feature`" + `).
3. Create ` + "`agentflow-init-<branch-slug>.md`" + ` (branch name with ` + "`/`" + ` replaced by ` + "`-`" + `, e.g. ` + "`agentflow-init-feature-my-feature.md`" + `) in the repository root (on ` + "`develop`" + `) with this content:

` + "```markdown" + `
# AgentFlow Init Request

## Parameters

- title: <feature title>
- branch: feature/<name>
- worktree: .worktrees/<name>
- root: develop

## Description

<2-3 sentences describing what this feature does and why>

## Scope

### In scope
- <item>

### Out of scope
- <item>

## Acceptance criteria
<checkable statements that define "done" — e.g. specific endpoint behavior,
test cases that must pass, error cases that must be handled. If you can't
write at least one, the scope probably isn't clear enough to start yet.>
- <criterion>

## Notes for CLI Agent
<context, constraints, or starting points the CLI Agent should know>
` + "```" + `

4. Tell the Human: *"` + "`agentflow-init-<branch-slug>.md`" + ` is ready. Ask the CLI Agent to run /agentflow-init."*

---

### Naming reference — branch-slug vs. feature-id

Two different names are derived from the branch, for two different purposes. Do not confuse them:

| Name           | Rule                                                                          | Used for                                  |
|----------------|--------------------------------------------------------------------------------|--------------------------------------------|
| ` + "`branch-slug`" + `  | The full branch name with every ` + "`/`" + ` replaced by ` + "`-`" + ` (nothing stripped)         | The init request filename                 |
| ` + "`feature-id`" + `   | Same as above, but a leading ` + "`feature/`" + ` prefix is dropped first                | The exchange folder under ` + "`.agentflow/features/`" + ` |

Example for branch ` + "`feature/my-feature`" + `:
- ` + "`branch-slug`" + ` = ` + "`feature-my-feature`" + ` → file ` + "`agentflow-init-feature-my-feature.md`" + `
- ` + "`feature-id`" + ` = ` + "`my-feature`" + ` → folder ` + "`.agentflow/features/my-feature/`" + `

The ` + "`feature/`" + ` prefix is dropped from ` + "`feature-id`" + ` because the parent directory is already named ` + "`features/`" + ` — keeping it would produce a redundant ` + "`.agentflow/features/feature-my-feature/`" + ` path.

---

### Fast track — small fixes and urgent bugs

Not every change needs discovery, plan, and implementation as three separate reviewed documents. Use the fast track only when **all** of these hold:

- The Human explicitly asks for it — you never decide to fast-track on your own.
- The fix is small and well understood: a handful of files, no schema/API/contract changes, nothing that needs discovery to even understand.
- Getting it wrong is low-risk and easily reversible (e.g. a revert).

How it differs from the normal flow:

1. You approve directly — set ` + "`STATUS.md`" + `: ` + "`Current phase: implementation`" + `, ` + "`Current turn: cli`" + `, ` + "`Next action: implement (fast track — no separate discovery/plan)`" + `.
2. The CLI Agent skips writing separate ` + "`discovery/`" + ` and ` + "`plans/`" + ` files, but still writes one ` + "`tasks/result-<date>.md`" + ` explaining what changed and why — same bar as a normal implementation review.
3. The core rule still applies without exception: **you review and approve the result before it's ` + "`done`" + `.** Only the discovery/plan documents are skipped, never your review.

If the CLI Agent discovers mid-fix that it's bigger than expected, it must stop and ask you to switch to the full flow instead of continuing to improvise on the fast track.

---

### Phase 1 — Discovery review

CLI Agent writes discovery results to ` + "`discovery/`" + `. Your job:

1. Read the discovery files on GitHub (feature branch).
2. Evaluate: complete and correct?
3. **Approve** → edit ` + "`STATUS.md`" + `: set turn to ` + "`cli`" + `, next action to ` + "`create_plan`" + `. Tell the Human.
   **Request changes** → create ` + "`discovery/feedback-<date>.md`" + ` with specific questions. Tell the Human.

---

### Phase 2 — Plan review

There is exactly one plan per feature: ` + "`plans/PLAN.md`" + `. The CLI Agent writes it once and, if scope evolves later, edits it in place and logs the change in its ` + "`## Changelog`" + ` section — it never creates a second plan file alongside it. Your job:

1. Read ` + "`plans/PLAN.md`" + `. Is the approach sound? Risks? Scope respected?
2. Check the plan against every acceptance criterion from the init request — does it address each one? A plan that doesn't mention a criterion isn't ready to approve.
3. If ` + "`plans/PLAN.md`" + ` already exists and you're reviewing it again (a revision, not the first version), read the ` + "`## Changelog`" + ` entry for what changed and review only that delta against the acceptance criteria it affects.
4. **Approve** → edit ` + "`STATUS.md`" + `: set turn to ` + "`cli`" + `, next action to ` + "`implement`" + `. Tell the Human.
   **Request changes** → create ` + "`plans/feedback-<date>.md`" + `, naming which acceptance criteria aren't covered. Tell the Human.

If new, separable work surfaces that isn't a revision of the current plan's scope, that's a signal for a new feature (` + "`agentflow init`" + ` again), not a second plan bolted onto this one.

---

### Phase 3 — Implementation review

CLI Agent writes task results to ` + "`tasks/`" + `. Your job:

1. Read the result files. Does the implementation match the approved plan?
2. Go through the acceptance criteria one by one — is each one actually met? If the result file doesn't say how a criterion was verified, ask before approving instead of assuming it was.
3. **Approve** → edit ` + "`STATUS.md`" + `: set status to ` + "`done`" + ` or next phase. Tell the Human.
   **Request changes** → create ` + "`reviews/feedback-<date>.md`" + `, naming which acceptance criteria failed or weren't verified. Tell the Human.

---

### Decision requests

When the CLI Agent creates a file in ` + "`decisions/`" + `, ` + "`STATUS.md`" + ` will show ` + "`current_turn: web`" + `.
Read it, discuss with the Human if needed, write the resolution back to that file, then return the turn to ` + "`cli`" + ` in ` + "`STATUS.md`" + `.

---

## How to update STATUS.md

Edit ` + "`STATUS.md`" + ` directly on GitHub (feature branch):

` + "```markdown" + `
## Current phase
<intake | discovery | planning | implementation | review | done>

## Current turn
<web | cli>

## Status
<short description>

## Next action
<what should happen next>

## Last update
<timestamp>
` + "```" + `

Also update ` + "`state.json`" + ` (same folder as ` + "`STATUS.md`" + `) with the same values, using this shape — change only ` + "`current_phase`" + `, ` + "`current_turn`" + `, ` + "`status`" + `, ` + "`next_action`" + `, and ` + "`updated_at`" + `; leave every other field as you found it:

` + "```json" + `
{
  "current_phase": "<intake | discovery | planning | implementation | review | done>",
  "current_turn": "<web | cli>",
  "status": "<short description>",
  "next_action": "<what should happen next>",
  "updated_at": "<UTC timestamp, e.g. 2026-07-05T20:34:57Z>"
}
` + "```" + `

` + "`current_phase`" + ` and ` + "`current_turn`" + ` use the exact same words in both files (` + "`intake`" + `, ` + "`web`" + `, etc.) — those two must match **literally**. ` + "`status`" + ` and ` + "`next_action`" + ` do not: ` + "`STATUS.md`" + ` holds a human-readable sentence (e.g. ` + "`Initialized.`" + `, ` + "`Read the exchange folder and begin Phase 1 — Discovery.`" + `) while ` + "`state.json`" + ` holds a short machine slug for the same fact (e.g. ` + "`initialized`" + `, ` + "`begin_discovery`" + `) — this is intentional, not a bug, so judge those two by meaning, never by exact text.

If ` + "`state.json`" + ` is missing, or ` + "`current_phase`" + `/` + "`current_turn`" + ` don't match literally, or ` + "`status`" + `/` + "`next_action`" + ` contradict each other in meaning (not just wording), stop and create a file in ` + "`decisions/`" + ` describing the mismatch instead of guessing which one is correct.

---

## What you must NOT do

- Do not edit source code files.
- Do not merge branches.
- Do not approve a phase you have not read.
- Do not skip discovery and go straight to planning.

---

## Communication pattern

Short, action-oriented messages to the Human:

- *"` + "`agentflow-init-<branch-slug>.md`" + ` is ready. Ask the CLI Agent to run /agentflow-init."*
- *"Plan approved. Ask the CLI Agent to implement."*
- *"Discovery needs more detail. Ask the CLI Agent to address the feedback."*
`
}

// ConfigMD returns the content for CONFIG.md.
func ConfigMD(featureID, title, root, branch, worktree, exchange string) string {
	return fmt.Sprintf(`# AgentFlow Config

feature_id: %s
title: %s
root_branch: %s
feature_branch: %s
worktree: %s
exchange_folder: %s

## Policies

- CLI must not implement before discovery and planning are approved.
- CLI must create result files at the end of each task.
- CLI must create decision requests when ambiguity or risk is found.
- CLI must update STATUS.md and state.json at every handoff.
- CLI may commit and push at the end of a completed turn.
`, featureID, title, root, branch, worktree, exchange)
}
