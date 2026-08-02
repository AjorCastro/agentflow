package protocol

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalRequiredDirs lists subdirectories that must exist under a local-mode
// coordination folder (.agentflow/local/<feature-id>/).
var LocalRequiredDirs = []string{"discussions", "history", "runtime"}

// LocalRequiredFiles lists files that must exist under a local-mode
// coordination folder as soon as it is initialized. task.md/result.md are
// intentionally excluded: they don't exist until the first round of work.
var LocalRequiredFiles = []string{"POLICY.md", "checkpoint.md"}

// ValidateLocal checks that all required files and dirs exist under localDir.
func ValidateLocal(localDir string) ValidationResult {
	result := ValidationResult{ExchangeFolder: localDir, Valid: true}

	for _, f := range LocalRequiredFiles {
		path := filepath.Join(localDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Missing = append(result.Missing, f)
			result.Valid = false
		}
	}

	for _, d := range LocalRequiredDirs {
		path := filepath.Join(localDir, d)
		if info, err := os.Stat(path); os.IsNotExist(err) || (err == nil && !info.IsDir()) {
			result.Missing = append(result.Missing, d+"/")
			result.Valid = false
		}
	}

	return result
}

// FindLocalFolders returns all directories matching .agentflow/local/*/ under root.
func FindLocalFolders(root string) ([]string, error) {
	pattern := filepath.Join(root, ".agentflow", "local", "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob error: %w", err)
	}

	var dirs []string
	for _, m := range matches {
		info, err := os.Stat(m)
		if err == nil && info.IsDir() {
			dirs = append(dirs, m)
		}
	}
	return dirs, nil
}

// DefaultLocalPath returns the default local-mode coordination folder path
// for a feature ID, mirroring DefaultExchangePath's convention for the
// GitHub-based exchange folder.
func DefaultLocalPath(repoDir, featureID string) string {
	return filepath.Join(repoDir, ".agentflow", "local", featureID)
}

// CurrentFeaturePath returns the path to the marker file that records which
// feature ID is "current" for a worktree — written by `agentflow local
// init`, read by every other `agentflow local` subcommand so the Human
// never has to pass --feature by hand once a feature is bootstrapped in a
// given worktree. Lives alongside the per-feature folders it points at, so
// it is excluded from Git by the same `.agentflow/local/` .gitignore entry.
func CurrentFeaturePath(repoDir string) string {
	return filepath.Join(repoDir, ".agentflow", "local", "CURRENT")
}

// WriteCurrentFeature records featureID as the current feature for repoDir.
func WriteCurrentFeature(repoDir, featureID string) error {
	path := CurrentFeaturePath(repoDir)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(featureID+"\n"), 0644)
}

// ReadCurrentFeature returns the feature ID last recorded by
// WriteCurrentFeature for repoDir, or an error if none was ever set.
func ReadCurrentFeature(repoDir string) (string, error) {
	data, err := os.ReadFile(CurrentFeaturePath(repoDir))
	if err != nil {
		return "", err
	}
	featureID := strings.TrimSpace(string(data))
	if featureID == "" {
		return "", fmt.Errorf("%s is empty", CurrentFeaturePath(repoDir))
	}
	return featureID, nil
}

// PolicyMD returns the initial content for POLICY.md: stable policies and
// metadata for a local-mode feature, written once at `agentflow local init`
// and updated only when a policy actually changes (never per-round, unlike
// checkpoint.md/task.md/result.md).
func PolicyMD(featureID, title, root, branch, worktree string) string {
	return fmt.Sprintf(`# AgentFlow Local — Policy — %s

feature_id: %s
title: %s
root_branch: %s
branch: %s
worktree: %s

## Roles

| Role      | Responsibility                                                        |
|-----------|------------------------------------------------------------------------|
| Human     | Signals turns between Controller and Coder; final decision-maker       |
| Controller| Talks to the Human, defines tasks, takes and records decisions, reviews results. Does not implement code. |
| Coder     | Investigates, plans, implements, tests. Starts a clean session per task. |

## Core rule

> Continuity lives in files (POLICY.md, checkpoint.md, task.md, result.md),
> never in conversation history. Both agents must be able to resume from a
> brand-new session reading only these artifacts.

## Validation gate

<Fill in with this repo's mandatory test/build commands and CI/promotion
gate, if any — see AGENTS.md/CONTRIBUTING.md at the repo root.>

## Commit conventions

<Fill in if this feature has commit-message or scope conventions beyond the
repo's defaults.>

## Never automatically read

- history/ — audit trail only, append-only, never a source of truth.
- runtime/ — session ids, PIDs, local transcripts; always discardable.
`, title, featureID, title, root, branch, worktree)
}

// CheckpointMD returns the initial (empty) content for checkpoint.md,
// written at `agentflow local init`. It is overwritten — never appended —
// at each milestone by whichever agent reaches one.
func CheckpointMD() string {
	return fmt.Sprintf(`# Checkpoint

## Goal and overall status
Feature just initialized. No task assigned yet.

## Last task and result
None yet.

## Files touched
None yet.

## Test/build status
Not run yet.

## Open blockers or risks
None yet.

## Non-obvious design decisions
None yet.

## Next suggested step
Controller: define the first task.md.

## Last update
%s
`, time.Now().UTC().Format(time.RFC3339))
}
