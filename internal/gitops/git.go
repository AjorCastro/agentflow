package gitops

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsGitRepo checks whether dir is inside a git repository.
func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// BranchExists returns true if the given branch ref exists locally.
func BranchExists(dir, branch string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", branch)
	return cmd.Run() == nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("could not determine current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ShortStatus returns the output of `git status --short`.
func ShortStatus(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "status", "--short").Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// AddWorktree adds a new worktree for an existing or new branch.
//
//	adopt == true  → branch already exists, just attach worktree
//	adopt == false → create a new branch from root
func AddWorktree(repoDir, worktreePath, branch, root string, adopt bool) error {
	var args []string
	if adopt {
		args = []string{"-C", repoDir, "worktree", "add", worktreePath, branch}
	} else {
		args = []string{"-C", repoDir, "worktree", "add", "-b", branch, worktreePath, root}
	}
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %w\n%s", err, out)
	}
	return nil
}

// CommitAll stages all changes inside worktreeDir and commits with the given message.
func CommitAll(worktreeDir, message string) error {
	addOut, err := exec.Command("git", "-C", worktreeDir, "add", "-A").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, addOut)
	}
	commitOut, err := exec.Command("git", "-C", worktreeDir, "commit", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, commitOut)
	}
	return nil
}

// HasRemote returns true if the repo at dir has at least one remote configured.
func HasRemote(dir string) bool {
	out, err := exec.Command("git", "-C", dir, "remote").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// ListWorktrees returns the paths of all worktrees linked to the repo at dir,
// excluding the main worktree itself.
func ListWorktrees(dir string) ([]string, error) {
	out, err := exec.Command("git", "-C", dir, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var paths []string
	first := true
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if first {
				first = false // skip main worktree
				continue
			}
			paths = append(paths, path)
		}
	}
	return paths, nil
}

// PushBranch pushes branch to origin, setting upstream.
func PushBranch(worktreeDir, branch string) error {
	out, err := exec.Command("git", "-C", worktreeDir, "push", "-u", "origin", branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, out)
	}
	return nil
}
