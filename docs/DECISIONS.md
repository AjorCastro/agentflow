# AgentFlow — Design Decisions

## DECISION-001: Use `os/exec` for Git, not a Go Git library

**Date:** 2026-06-27  
**Status:** Accepted

**Context:** Several Go Git libraries exist (go-git, libgit2 bindings). The alternative was to use `os/exec` to call the system `git` binary.

**Decision:** Use `os/exec` with the system `git` binary.

**Rationale:**
- Git worktree operations (`git worktree add`) are complex and not fully supported in all Go Git libraries at parity with CLI.
- The system git respects local config, hooks, credentials, and SSH agents automatically.
- The dependency footprint stays minimal (zero extra deps for Git).
- Error messages from real git are more actionable for users than library-level errors.

**Trade-offs:** Requires git to be installed on the system. Not portable to environments without git CLI.

---

## DECISION-002: Exchange folder defaults to `.agentflow/features/<feature-id>/`

**Date:** 2026-06-27  
**Status:** Accepted

**Context:** The exchange folder must be discoverable by all roles without configuration. It must live on the feature branch.

**Decision:** Default to `.agentflow/features/<feature-id>/` relative to the worktree root. `feature-id` is derived from the branch name by replacing `/` with `-`.

**Rationale:**
- Glob pattern `.agentflow/features/*/` makes multi-feature discovery trivial.
- Derived from branch name means no extra flag needed in the common case.
- The `.agentflow/` prefix namespaces away from project files.

---

## DECISION-003: `--worktree` is required by default; `--no-worktree` as an escape hatch

**Date:** 2026-06-27  
**Status:** Accepted

**Context:** In the standard flow the CLI agent works in an isolated worktree to avoid interfering with the main checkout. But some users may want to operate on a single checkout.

**Decision:** `--worktree` is required unless `--no-worktree` is set. With `--no-worktree`, the CLI fails if the repo is not already on the target branch (no implicit checkout).

**Rationale:**
- Implicit `git checkout` is destructive and surprising. Explicit fail is safer.
- The escape hatch covers single-checkout environments (CI, simple repos) without making it the default.

---

## DECISION-004: No TUI, no interactive prompts

**Date:** 2026-06-27  
**Status:** Accepted

**Context:** Interactive prompts (survey, bubbletea) could improve ergonomics for first-time users.

**Decision:** The CLI is purely flag-driven with no interactive prompts.

**Rationale:**
- `agentflow init` is typically called from a script or from Web Reviewer instructions. Flags compose better with scripts and automation.
- Keeps the dependency footprint minimal.
- TUI can be added later as a separate `agentflow tui` subcommand without breaking existing usage.

---

## DECISION-005: state.json stores `worktree` and `exchange_folder` as relative paths

**Date:** 2026-06-27  
**Status:** Open / Under review

**Context:** Absolute paths in state.json break portability if the repo is moved or shared.

**Decision (current):** Store whatever the user passed to `--worktree` and `--exchange`. If no-worktree mode is used, store `"(none — operating on repo)"` for worktree.

**Open question:** Should we normalize to relative-from-repo paths on write? This would require passing the repo root to the state writer. Deferred to a future iteration.
