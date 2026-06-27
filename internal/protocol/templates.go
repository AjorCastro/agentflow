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
