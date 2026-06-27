# AgentFlow Exchange — Test init local

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

feature_id: feature-test-init
title: Test init local
