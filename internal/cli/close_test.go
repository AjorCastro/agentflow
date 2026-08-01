package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AjorCastro/agentflow/internal/protocol"
)

// runGit runs a git command in dir, failing the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// newMergedFeatureRepo creates a scratch repo on branch "develop" with a
// feature branch already merged in, with no worktree ever attached to it.
func newMergedFeatureRepo(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()

	runGit(t, dir, "init", "-q", "-b", "develop")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("root\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-q", "-m", "init")

	runGit(t, dir, "checkout", "-q", "-b", branch)
	if err := os.WriteFile(filepath.Join(dir, "feature.txt"), []byte("feature\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "feature.txt")
	runGit(t, dir, "commit", "-q", "-m", "feature work")

	runGit(t, dir, "checkout", "-q", "develop")
	runGit(t, dir, "merge", "-q", "--no-ff", "-m", "merge feature", branch)

	return dir
}

// TestRunClose_RemovesLocalDirWithoutWorktree covers the gap identified in
// task.md: when local mode was initialized directly on a repo checkout
// (no dedicated worktree for the feature branch), nothing removes
// .agentflow/local/<feature-id> other than the explicit cleanup in
// runClose — git worktree remove never runs in this case since there is no
// linked worktree to remove.
func TestRunClose_RemovesLocalDirWithoutWorktree(t *testing.T) {
	branch := "feature/local-close-test"
	dir := newMergedFeatureRepo(t, branch)

	featureID := protocol.FeatureIDFromBranch(branch)
	localDir := protocol.DefaultLocalPath(dir, featureID)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "checkpoint.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := &closeOptions{repo: dir, root: "develop", branch: branch}
	if err := runClose(opts); err != nil {
		t.Fatalf("runClose failed: %v", err)
	}

	if _, err := os.Stat(localDir); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, stat err = %v", localDir, err)
	}
}

// TestRunClose_NoLocalDir verifies runClose is a no-op regarding local-mode
// cleanup when no local-mode folder was ever created for the feature.
func TestRunClose_NoLocalDir(t *testing.T) {
	branch := "feature/no-local-dir"
	dir := newMergedFeatureRepo(t, branch)

	opts := &closeOptions{repo: dir, root: "develop", branch: branch}
	if err := runClose(opts); err != nil {
		t.Fatalf("runClose failed: %v", err)
	}
}
