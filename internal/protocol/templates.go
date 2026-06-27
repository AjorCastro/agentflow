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
| plans/         | Implementation plans waiting for approval            |
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
human

## Status
Initialized.

## Next action
Define the feature and ask the CLI Agent to run the initial AgentFlow skill.

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

## Your tasks by phase

### Phase 0 — Feature definition (before agentflow init)

The Human describes what they want to build. Your job:

1. Ask clarifying questions until you understand scope and constraints.
2. Propose: title, branch name (e.g. ` + "`feature/my-feature`" + `), worktree path (e.g. ` + "`.worktrees/my-feature`" + `).
3. Create ` + "`agentflow-init.md`" + ` in the repository root (on ` + "`develop`" + `) with this content:

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

## Notes for CLI Agent
<context, constraints, or starting points the CLI Agent should know>
` + "```" + `

4. Tell the Human: *"agentflow-init.md is ready. Ask the CLI Agent to run /agentflow-init."*

---

### Phase 1 — Discovery review

CLI Agent writes discovery results to ` + "`discovery/`" + `. Your job:

1. Read the discovery files on GitHub (feature branch).
2. Evaluate: complete and correct?
3. **Approve** → edit ` + "`STATUS.md`" + `: set turn to ` + "`cli`" + `, next action to ` + "`create_plan`" + `. Tell the Human.
   **Request changes** → create ` + "`discovery/feedback-<date>.md`" + ` with specific questions. Tell the Human.

---

### Phase 2 — Plan review

CLI Agent writes a plan to ` + "`plans/`" + `. Your job:

1. Read the plan. Is the approach sound? Risks? Scope respected?
2. **Approve** → edit ` + "`STATUS.md`" + `: set turn to ` + "`cli`" + `, next action to ` + "`implement`" + `. Tell the Human.
   **Request changes** → create ` + "`plans/feedback-<date>.md`" + `. Tell the Human.

---

### Phase 3 — Implementation review

CLI Agent writes task results to ` + "`tasks/`" + `. Your job:

1. Read the result files. Does the implementation match the approved plan?
2. **Approve** → edit ` + "`STATUS.md`" + `: set status to ` + "`done`" + ` or next phase. Tell the Human.
   **Request changes** → create ` + "`reviews/feedback-<date>.md`" + `. Tell the Human.

---

### Decision requests

When the CLI Agent creates a file in ` + "`decisions/`" + "`, `STATUS.md`" + ` will show ` + "`current_turn: human`" + `.
Read it, discuss with the Human, write the resolution back to that file, then return the turn to ` + "`cli`" + ` in ` + "`STATUS.md`" + `.

---

## How to update STATUS.md

Edit ` + "`STATUS.md`" + ` directly on GitHub (feature branch):

` + "```markdown" + `
## Current phase
<intake | discovery | planning | implementation | review | done>

## Current turn
<human | web | cli>

## Status
<short description>

## Next action
<what should happen next>

## Last update
<timestamp>
` + "```" + `

Also update ` + "`state.json`" + ` with the same values.

---

## What you must NOT do

- Do not edit source code files.
- Do not merge branches.
- Do not approve a phase you have not read.
- Do not skip discovery and go straight to planning.

---

## Communication pattern

Short, action-oriented messages to the Human:

- *"agentflow-init.md is ready. Ask the CLI Agent to run /agentflow-init."*
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
