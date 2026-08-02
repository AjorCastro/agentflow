package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AjorCastro/agentflow/internal/protocol"
	"github.com/spf13/cobra"
)

// newGitRepo creates a scratch git repository (no local-mode folder yet) for
// tests that only need a valid git repo to point --repo at. Reuses the
// runGit helper defined in close_test.go (same package).
func newGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("root\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-q", "-m", "init")

	return dir
}

// initLocal runs `agentflow local init` for repo/feature and fails the test
// on error, for tests that need an already-initialized local-mode folder.
func initLocal(t *testing.T, repo, feature string) {
	t.Helper()
	cmd := newLocalInitCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", feature})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("local init failed: %v", err)
	}
}

// writeTempFile writes content to a new file in a fresh temp dir and
// returns its path, for testing the --file flag.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "content.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// withStdin temporarily replaces os.Stdin with a pipe fed with content,
// runs fn, then restores it. Needed because readContent (local.go) reads
// os.Stdin directly rather than cmd.InOrStdin(), so cobra's SetIn does not
// intercept it.
func withStdin(t *testing.T, content string, fn func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = w.WriteString(content)
		_ = w.Close()
	}()
	fn()
	<-done
}

// captureStdout redirects os.Stdout to a pipe for the duration of fn and
// returns everything written to it. Needed because local.go's status/
// context/discuss commands print via fmt.Printf directly to os.Stdout
// rather than cmd.OutOrStdout(), so cobra's SetOut does not intercept it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan string)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	fn()
	_ = w.Close()
	return <-done
}

func TestLocalInit_Success(t *testing.T) {
	repo := newGitRepo(t)

	cmd := newLocalInitCmd()
	cmd.SetArgs([]string{
		"--repo", repo,
		"--feature", "my-feature",
		"--title", "My Feature",
		"--root", "develop",
		"--branch", "feature/my-feature",
		"--worktree", ".worktrees/my-feature",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	localDir := filepath.Join(repo, ".agentflow", "local", "my-feature")

	policy, err := os.ReadFile(filepath.Join(localDir, "POLICY.md"))
	if err != nil {
		t.Fatalf("reading POLICY.md: %v", err)
	}
	for _, want := range []string{
		"feature_id: my-feature",
		"title: My Feature",
		"root_branch: develop",
		"branch: feature/my-feature",
		"worktree: .worktrees/my-feature",
	} {
		if !strings.Contains(string(policy), want) {
			t.Errorf("POLICY.md missing %q\ngot:\n%s", want, policy)
		}
	}

	checkpoint, err := os.ReadFile(filepath.Join(localDir, "checkpoint.md"))
	if err != nil {
		t.Fatalf("reading checkpoint.md: %v", err)
	}
	if !strings.Contains(string(checkpoint), "# Checkpoint") {
		t.Errorf("checkpoint.md missing header, got:\n%s", checkpoint)
	}

	for _, dir := range append([]string{"."}, protocol.LocalRequiredDirs...) {
		info, err := os.Stat(filepath.Join(localDir, dir))
		if err != nil || !info.IsDir() {
			t.Errorf("expected dir %s to exist, err=%v", dir, err)
		}
	}
}

func TestLocalInit_FailsIfAlreadyExists(t *testing.T) {
	repo := newGitRepo(t)
	args := []string{"--repo", repo, "--feature", "dup-feature"}

	first := newLocalInitCmd()
	first.SetArgs(args)
	if err := first.Execute(); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	second := newLocalInitCmd()
	second.SetArgs(args)
	err := second.Execute()
	if err == nil {
		t.Fatal("expected error on second init, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestLocalInit_FailsWithoutFeature(t *testing.T) {
	repo := newGitRepo(t)

	cmd := newLocalInitCmd()
	cmd.SetArgs([]string{"--repo", repo})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when --feature is missing")
	}
}

func TestLocalInit_FailsIfNotGitRepo(t *testing.T) {
	dir := t.TempDir() // not a git repo

	cmd := newLocalInitCmd()
	cmd.SetArgs([]string{"--repo", dir, "--feature", "no-git"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-git repo")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("expected 'not a git repository' error, got: %v", err)
	}
}

func TestLocalRoundFile_WritesFromStdin(t *testing.T) {
	cases := []struct {
		name     string
		newCmd   func() *cobra.Command
		filename string
	}{
		{"task", newLocalTaskCmd, "task.md"},
		{"result", newLocalResultCmd, "result.md"},
		{"checkpoint", newLocalCheckpointCmd, "checkpoint.md"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newGitRepo(t)
			feature := "feat-" + tc.name
			initLocal(t, repo, feature)

			content := "# " + tc.name + " content\n"
			cmd := tc.newCmd()
			cmd.SetArgs([]string{"--repo", repo, "--feature", feature})

			withStdin(t, content, func() {
				if err := cmd.Execute(); err != nil {
					t.Fatalf("%s failed: %v", tc.name, err)
				}
			})

			localDir := filepath.Join(repo, ".agentflow", "local", feature)
			got, err := os.ReadFile(filepath.Join(localDir, tc.filename))
			if err != nil {
				t.Fatalf("reading %s: %v", tc.filename, err)
			}
			if string(got) != content {
				t.Errorf("got %q, want %q", got, content)
			}
		})
	}
}

func TestLocalRoundFile_WritesFromFile(t *testing.T) {
	repo := newGitRepo(t)
	initLocal(t, repo, "feat-file")

	content := "# from file\n"
	srcPath := writeTempFile(t, content)

	cmd := newLocalTaskCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", "feat-file", "--file", srcPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("task --file failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(repo, ".agentflow", "local", "feat-file", "task.md"))
	if err != nil {
		t.Fatalf("reading task.md: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLocalRoundFile_FailsWithoutInit(t *testing.T) {
	repo := newGitRepo(t) // local-mode folder never initialized

	srcPath := writeTempFile(t, "content\n")

	cmd := newLocalTaskCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", "never-initialized", "--file", srcPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when local dir doesn't exist")
	}
	if !strings.Contains(err.Error(), "run `agentflow local init` first") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalRoundFile_ArchivesPreviousOnOverwrite(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-archive"
	initLocal(t, repo, feature)
	localDir := filepath.Join(repo, ".agentflow", "local", feature)

	// task.md does not exist after init, so the first write should not
	// archive anything.
	first := newLocalTaskCmd()
	first.SetArgs([]string{"--repo", repo, "--feature", feature})
	withStdin(t, "v1\n", func() {
		if err := first.Execute(); err != nil {
			t.Fatalf("first task write failed: %v", err)
		}
	})

	entries, err := os.ReadDir(filepath.Join(localDir, "history"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no archived files after first write, got %d", len(entries))
	}

	second := newLocalTaskCmd()
	second.SetArgs([]string{"--repo", repo, "--feature", feature})
	withStdin(t, "v2\n", func() {
		if err := second.Execute(); err != nil {
			t.Fatalf("second task write failed: %v", err)
		}
	})

	got, err := os.ReadFile(filepath.Join(localDir, "task.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v2\n" {
		t.Errorf("expected current task.md = v2, got %q", got)
	}

	entries, err = os.ReadDir(filepath.Join(localDir, "history"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 archived file after second write, got %d: %v", len(entries), entries)
	}
	if !strings.HasSuffix(entries[0].Name(), "-task.md") {
		t.Errorf("expected archived filename to end with -task.md, got %q", entries[0].Name())
	}

	archived, err := os.ReadFile(filepath.Join(localDir, "history", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if string(archived) != "v1\n" {
		t.Errorf("expected archived content = v1, got %q", archived)
	}
}

func TestArchiveAndWrite_ArchivesPrevious(t *testing.T) {
	localDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(localDir, "history"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := archiveAndWrite(localDir, "checkpoint.md", "first version\n"); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(localDir, "history"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no archived files after first write (nothing to archive), got %d", len(entries))
	}

	if err := archiveAndWrite(localDir, "checkpoint.md", "second version\n"); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	current, err := os.ReadFile(filepath.Join(localDir, "checkpoint.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "second version\n" {
		t.Errorf("expected current file to hold second version, got %q", current)
	}

	entries, err = os.ReadDir(filepath.Join(localDir, "history"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 archived file, got %d: %v", len(entries), entries)
	}

	archivedName := entries[0].Name()
	if !strings.HasSuffix(archivedName, "-checkpoint.md") {
		t.Errorf("expected archived filename to end with -checkpoint.md, got %q", archivedName)
	}

	archivedContent, err := os.ReadFile(filepath.Join(localDir, "history", archivedName))
	if err != nil {
		t.Fatal(err)
	}
	if string(archivedContent) != "first version\n" {
		t.Errorf("expected archived content to be first version, got %q", archivedContent)
	}
}

func TestLocalDiscussStart_CreatesIncrementingIDs(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-discuss"
	initLocal(t, repo, feature)
	localDir := filepath.Join(repo, ".agentflow", "local", feature)

	first := newLocalDiscussStartCmd()
	first.SetArgs([]string{"--repo", repo, "--feature", feature})
	out := captureStdout(t, func() {
		if err := first.Execute(); err != nil {
			t.Fatalf("first discuss start failed: %v", err)
		}
	})
	if !strings.Contains(out, "Discussion 001 started") {
		t.Errorf("expected output to mention discussion 001, got: %q", out)
	}
	if info, err := os.Stat(filepath.Join(localDir, "discussions", "001")); err != nil || !info.IsDir() {
		t.Errorf("expected discussions/001 to exist, err=%v", err)
	}

	second := newLocalDiscussStartCmd()
	second.SetArgs([]string{"--repo", repo, "--feature", feature})
	out = captureStdout(t, func() {
		if err := second.Execute(); err != nil {
			t.Fatalf("second discuss start failed: %v", err)
		}
	})
	if !strings.Contains(out, "Discussion 002 started") {
		t.Errorf("expected output to mention discussion 002, got: %q", out)
	}
	if info, err := os.Stat(filepath.Join(localDir, "discussions", "002")); err != nil || !info.IsDir() {
		t.Errorf("expected discussions/002 to exist, err=%v", err)
	}
}

func TestLocalDiscussStart_FailsWithoutInit(t *testing.T) {
	repo := newGitRepo(t) // local-mode folder never initialized

	cmd := newLocalDiscussStartCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", "never-initialized"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when local dir doesn't exist")
	}
	if !strings.Contains(err.Error(), "run `agentflow local init` first") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalDiscussClose_FromStdin(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-discuss-close-stdin"
	initLocal(t, repo, feature)

	start := newLocalDiscussStartCmd()
	start.SetArgs([]string{"--repo", repo, "--feature", feature})
	captureStdout(t, func() {
		if err := start.Execute(); err != nil {
			t.Fatalf("discuss start failed: %v", err)
		}
	})

	content := "# Outcome\nDecided X.\n"
	closeCmd := newLocalDiscussCloseCmd()
	closeCmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--id", "001"})
	withStdin(t, content, func() {
		if err := closeCmd.Execute(); err != nil {
			t.Fatalf("discuss close failed: %v", err)
		}
	})

	localDir := filepath.Join(repo, ".agentflow", "local", feature)
	got, err := os.ReadFile(filepath.Join(localDir, "discussions", "001", "OUTCOME.md"))
	if err != nil {
		t.Fatalf("reading OUTCOME.md: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLocalDiscussClose_FromFile(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-discuss-close-file"
	initLocal(t, repo, feature)

	start := newLocalDiscussStartCmd()
	start.SetArgs([]string{"--repo", repo, "--feature", feature})
	captureStdout(t, func() {
		if err := start.Execute(); err != nil {
			t.Fatalf("discuss start failed: %v", err)
		}
	})

	content := "# Outcome from file\n"
	srcPath := writeTempFile(t, content)

	closeCmd := newLocalDiscussCloseCmd()
	closeCmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--id", "001", "--file", srcPath})
	if err := closeCmd.Execute(); err != nil {
		t.Fatalf("discuss close --file failed: %v", err)
	}

	localDir := filepath.Join(repo, ".agentflow", "local", feature)
	got, err := os.ReadFile(filepath.Join(localDir, "discussions", "001", "OUTCOME.md"))
	if err != nil {
		t.Fatalf("reading OUTCOME.md: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLocalDiscussClose_FailsIfDiscussionDoesNotExist(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-discuss-close-missing"
	initLocal(t, repo, feature)

	srcPath := writeTempFile(t, "content\n")

	closeCmd := newLocalDiscussCloseCmd()
	closeCmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--id", "999", "--file", srcPath})

	err := closeCmd.Execute()
	if err == nil {
		t.Fatal("expected error when discussion id doesn't exist")
	}
	if !strings.Contains(err.Error(), "run `agentflow local discuss start` first") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalDiscussClose_FailsWithoutInit(t *testing.T) {
	repo := newGitRepo(t) // local-mode folder never initialized

	srcPath := writeTempFile(t, "content\n")

	closeCmd := newLocalDiscussCloseCmd()
	closeCmd.SetArgs([]string{"--repo", repo, "--feature", "never-initialized", "--id", "001", "--file", srcPath})

	err := closeCmd.Execute()
	if err == nil {
		t.Fatal("expected error when local dir doesn't exist")
	}
}

func TestLocalStatus_Success(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-status"
	initLocal(t, repo, feature)

	// Write task.md so status has a mix of written/unwritten files to report.
	taskCmd := newLocalTaskCmd()
	taskCmd.SetArgs([]string{"--repo", repo, "--feature", feature})
	withStdin(t, "# Task\n", func() {
		if err := taskCmd.Execute(); err != nil {
			t.Fatalf("task write failed: %v", err)
		}
	})

	cmd := newLocalStatusCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", feature})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("status failed: %v", err)
		}
	})

	if !strings.Contains(out, "task.md") {
		t.Errorf("expected status output to mention task.md, got:\n%s", out)
	}
	if !strings.Contains(out, "not written yet") {
		t.Errorf("expected status output to flag unwritten files (e.g. result.md), got:\n%s", out)
	}
	if !strings.Contains(out, "discussions/") {
		t.Errorf("expected status output to mention discussions/, got:\n%s", out)
	}
}

func TestLocalStatus_FailsWithoutInit(t *testing.T) {
	repo := newGitRepo(t) // local-mode folder never initialized

	cmd := newLocalStatusCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", "never-initialized"})

	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("status should not error for a missing folder, it should report Missing: %v", err)
		}
	})
	if !strings.Contains(out, "Missing:") {
		t.Errorf("expected status output to report Missing:, got:\n%s", out)
	}
}

func TestLocalContext_CoderExcludesResultAndAuditDirs(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-context-coder"
	initLocal(t, repo, feature)
	localDir := filepath.Join(repo, ".agentflow", "local", feature)

	// Write task.md and result.md so we can confirm the coder role excludes
	// result.md even though it exists.
	for _, name := range []string{"task.md", "result.md"} {
		if err := os.WriteFile(filepath.Join(localDir, name), []byte("# "+name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// history/ and runtime/ must never be read automatically, even if they
	// contain files — plant one in each to prove context ignores them.
	if err := os.WriteFile(filepath.Join(localDir, "history", "secret-history.md"), []byte("SECRET-HISTORY-CONTENT"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "runtime", "secret-runtime.md"), []byte("SECRET-RUNTIME-CONTENT"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newLocalContextCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--role", "coder"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("context failed: %v", err)
		}
	})

	for _, want := range []string{"--- POLICY.md ---", "--- checkpoint.md ---", "--- task.md ---"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected coder context output to include %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "--- result.md ---") {
		t.Errorf("coder context must not include result.md, got:\n%s", out)
	}
	if strings.Contains(out, "SECRET-HISTORY-CONTENT") || strings.Contains(out, "SECRET-RUNTIME-CONTENT") {
		t.Errorf("coder context must never include history/ or runtime/ content, got:\n%s", out)
	}
}

func TestLocalContext_ControllerIncludesResult(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-context-controller"
	initLocal(t, repo, feature)
	localDir := filepath.Join(repo, ".agentflow", "local", feature)

	for _, name := range []string{"task.md", "result.md"} {
		if err := os.WriteFile(filepath.Join(localDir, name), []byte("# "+name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(localDir, "history", "secret-history.md"), []byte("SECRET-HISTORY-CONTENT"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "runtime", "secret-runtime.md"), []byte("SECRET-RUNTIME-CONTENT"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newLocalContextCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--role", "controller"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("context failed: %v", err)
		}
	})

	for _, want := range []string{"--- POLICY.md ---", "--- checkpoint.md ---", "--- task.md ---", "--- result.md ---"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected controller context output to include %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "SECRET-HISTORY-CONTENT") || strings.Contains(out, "SECRET-RUNTIME-CONTENT") {
		t.Errorf("controller context must never include history/ or runtime/ content, got:\n%s", out)
	}
}

func TestLocalContext_SkipsMissingFiles(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-context-missing"
	initLocal(t, repo, feature) // task.md/result.md not written yet

	cmd := newLocalContextCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--role", "controller"})
	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("context failed: %v", err)
		}
	})

	for _, want := range []string{"--- POLICY.md ---", "--- checkpoint.md ---"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to include %q, got:\n%s", want, out)
		}
	}
	for _, notWant := range []string{"--- task.md ---", "--- result.md ---"} {
		if strings.Contains(out, notWant) {
			t.Errorf("expected output to skip missing %q, got:\n%s", notWant, out)
		}
	}
}

func TestLocalContext_InvalidRole(t *testing.T) {
	repo := newGitRepo(t)
	feature := "feat-context-bad-role"
	initLocal(t, repo, feature)

	cmd := newLocalContextCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", feature, "--role", "bogus"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --role")
	}
	if !strings.Contains(err.Error(), "controller") || !strings.Contains(err.Error(), "coder") {
		t.Errorf("expected error to mention both valid roles, got: %v", err)
	}
}

func TestLocalInit_RecordsCurrentFeature(t *testing.T) {
	repo := newGitRepo(t)
	initLocal(t, repo, "auto-linked")

	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		t.Fatal(err)
	}
	got, err := protocol.ReadCurrentFeature(repoAbs)
	if err != nil {
		t.Fatalf("reading current feature: %v", err)
	}
	if got != "auto-linked" {
		t.Errorf("expected current feature %q, got %q", "auto-linked", got)
	}
}

func TestLocalRoundFile_FallsBackToCurrentFeature(t *testing.T) {
	repo := newGitRepo(t)
	initLocal(t, repo, "auto-linked")

	cmd := newLocalTaskCmd()
	cmd.SetArgs([]string{"--repo", repo}) // no --feature
	withStdin(t, "# Task\nauto-linked feature\n", func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("task without --feature failed: %v", err)
		}
	})

	got, err := os.ReadFile(filepath.Join(repo, ".agentflow", "local", "auto-linked", "task.md"))
	if err != nil {
		t.Fatalf("reading task.md: %v", err)
	}
	if !strings.Contains(string(got), "auto-linked feature") {
		t.Errorf("task.md missing expected content, got:\n%s", got)
	}
}

func TestLocalRoundFile_ExplicitFeatureOverridesCurrent(t *testing.T) {
	repo := newGitRepo(t)
	initLocal(t, repo, "feature-a")
	initLocal(t, repo, "feature-b") // becomes the current feature

	cmd := newLocalTaskCmd()
	cmd.SetArgs([]string{"--repo", repo, "--feature", "feature-a"})
	withStdin(t, "for feature-a\n", func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("task with explicit --feature failed: %v", err)
		}
	})

	if _, err := os.Stat(filepath.Join(repo, ".agentflow", "local", "feature-b", "task.md")); err == nil {
		t.Error("task.md should not have been written under the current (feature-b), unrelated feature")
	}
	got, err := os.ReadFile(filepath.Join(repo, ".agentflow", "local", "feature-a", "task.md"))
	if err != nil {
		t.Fatalf("reading task.md: %v", err)
	}
	if !strings.Contains(string(got), "for feature-a") {
		t.Errorf("task.md missing expected content, got:\n%s", got)
	}
}

func TestLocalRoundFile_FailsWithoutFeatureOrCurrent(t *testing.T) {
	repo := newGitRepo(t) // no `local init` ever run, so no current feature

	cmd := newLocalTaskCmd()
	cmd.SetArgs([]string{"--repo", repo})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	withStdin(t, "content\n", func() {
		if err := cmd.Execute(); err == nil {
			t.Fatal("expected error when neither --feature nor a current feature is set")
		} else if !strings.Contains(err.Error(), "no --feature given") {
			t.Errorf("expected 'no --feature given' error, got: %v", err)
		}
	})
}
