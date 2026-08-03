package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AjorCastro/agentflow/internal/gitops"
	"github.com/AjorCastro/agentflow/internal/protocol"
	"github.com/spf13/cobra"
)

// newLocalCmd registers the `agentflow local` command group: the prototype
// v0 CLI surface for local mode (Controller + Coder coordinating through
// small files instead of a GitHub-based exchange folder). See
// .agentflow-local/discussions/001/PROTOTYPE-PLAN.md for the design this
// implements.
func newLocalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Local mode: coordinate Controller and Coder via files instead of GitHub",
		Long: `Local mode coordinates two agent sessions — a Controller (talks to you,
assigns tasks, reviews results, never writes code) and a Coder (investigates,
implements, tests) — through a handful of small files instead of a GitHub
exchange folder. See docs/LOCAL-MODE-GUIDE.md for the full walkthrough.

Once a feature is bootstrapped, every command below auto-detects which
feature you mean from this worktree's current feature (set by
'agentflow local init') — --feature is only needed to target a different one.

Quick start:

  1. Decide the feature with the Human, then bootstrap it — in a Controller
     session, invoke the agentflow-local-init skill. It runs in two stages:

     Stage A (pre-worktree, from the root repo checkout): bootstraps a
     pending coordination folder, analyzes the problem/need with the Human,
     assesses impact via a sub-agent, drafts and validates PLAN.md, and gets
     the Human's go/no-go:
       agentflow local init --repo <root-repo-path> --feature <name>
       agentflow local plan <<'EOF' ... EOF

     Stage B (after approval): creates the worktree, promotes the pending
     folder into it, fills in POLICY.md, and seeds the first task.md from
     PLAN.md's first checklist item:
       agentflow local promote --feature <name> --from <root-repo-path>
       agentflow local task <<'EOF' ... EOF

     A Human-requested fast track skips Stage A entirely for small,
     well-understood, low-risk fixes.

  2. In a second, independent session opened in the same worktree, invoke
     the agentflow-local-coder-init skill. It confirms the link and loads
     only POLICY.md+checkpoint.md+task.md via:
       agentflow local context --role coder
     then reports back with:
       agentflow local result <<'EOF' ... EOF

  3. Back in the Controller session, review result.md, check off the
     completed PLAN.md item, assign the next task or close the feature.
     Check progress any time with:
       agentflow local status
     (shows PLAN.md's checklist progress as "N/M done" when one exists)

  4. For every session after the first one, use agentflow-local-resume
     (Controller) or agentflow-local-coder-resume (Coder) instead of the
     -init skills — including after a forced session restart mid-feature.
     PLAN.md (not just checkpoint.md) is what keeps the work sequence intact
     across those restarts.

  5. If the Human needs to interrupt a session off a natural milestone (end
     of day, an unrelated interruption), use the agentflow-local-pause skill
     first so the next resume doesn't start blind.

  6. When every PLAN.md item is done (or the fast-tracked task is done),
     close the feature exactly like GitHub mode:
       agentflow close --branch feature/<name> --root develop
     This also removes .agentflow/local/<name>/ — nothing to clean up by hand.`,
	}

	cmd.AddCommand(newLocalInitCmd())
	cmd.AddCommand(newLocalPlanCmd())
	cmd.AddCommand(newLocalPromoteCmd())
	cmd.AddCommand(newLocalTaskCmd())
	cmd.AddCommand(newLocalResultCmd())
	cmd.AddCommand(newLocalCheckpointCmd())
	cmd.AddCommand(newLocalDiscussCmd())
	cmd.AddCommand(newLocalStatusCmd())
	cmd.AddCommand(newLocalContextCmd())

	return cmd
}

// resolveFeature returns the feature ID a command should operate on:
// explicit if --feature was passed, otherwise whatever `agentflow local
// init` last recorded as current for this worktree. This is what lets every
// subcommand except `init` itself omit --feature once a feature exists.
func resolveFeature(repoAbs, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	current, err := protocol.ReadCurrentFeature(repoAbs)
	if err != nil {
		return "", fmt.Errorf("no --feature given and no current feature set in %s — pass --feature explicitly or run `agentflow local init` first", repoAbs)
	}
	return current, nil
}

// resolveLocalDir returns the absolute local-mode coordination folder for
// --repo/--feature (or the worktree's current feature if --feature is
// omitted), validating repo is a git repository.
func resolveLocalDir(repo, feature string) (string, error) {
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return "", fmt.Errorf("invalid --repo path: %w", err)
	}
	if !gitops.IsGitRepo(repoAbs) {
		return "", fmt.Errorf("%s is not a git repository", repoAbs)
	}
	featureID, err := resolveFeature(repoAbs, feature)
	if err != nil {
		return "", err
	}
	return protocol.DefaultLocalPath(repoAbs, featureID), nil
}

// readContent returns the new content for a round-exchange file: from
// --file if set, otherwise from stdin. Stdin/heredoc is the primary
// contract (see PROTOTYPE-PLAN.md §9) because these commands are meant to
// be invoked by an LLM agent via its shell tool, not typed interactively.
func readContent(file string) (string, error) {
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("reading --file: %w", err)
		}
		return string(data), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("no content provided: pipe content via stdin or pass --file")
	}
	return string(data), nil
}

// archiveAndWrite writes content to localDir/name, first copying any
// existing file at that path into history/ with a timestamped name. history/
// is audit-only: it is never read back automatically by either agent.
func archiveAndWrite(localDir, name, content string) error {
	target := filepath.Join(localDir, name)

	if existing, err := os.ReadFile(target); err == nil {
		stamp := time.Now().UTC().Format("20060102T150405Z")
		archived := filepath.Join(localDir, "history", fmt.Sprintf("%s-%s", stamp, name))
		if err := os.WriteFile(archived, existing, 0644); err != nil {
			return fmt.Errorf("archiving previous %s: %w", name, err)
		}
	}

	return os.WriteFile(target, []byte(content), 0644)
}

func newLocalInitCmd() *cobra.Command {
	var repo, feature, title, root, branch, worktree string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the local-mode coordination folder for a feature",
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir, err := resolveLocalDir(repo, feature)
			if err != nil {
				return err
			}
			if title == "" {
				title = feature
			}

			if _, err := os.Stat(localDir); err == nil {
				return fmt.Errorf("%s already exists", localDir)
			}

			for _, sub := range append([]string{"."}, protocol.LocalRequiredDirs...) {
				if err := os.MkdirAll(filepath.Join(localDir, sub), 0755); err != nil {
					return err
				}
			}

			files := map[string]string{
				"POLICY.md":     protocol.PolicyMD(feature, title, root, branch, worktree),
				"checkpoint.md": protocol.CheckpointMD(),
			}
			for name, content := range files {
				if err := os.WriteFile(filepath.Join(localDir, name), []byte(content), 0644); err != nil {
					return err
				}
			}

			repoAbs, err := filepath.Abs(repo)
			if err != nil {
				return fmt.Errorf("invalid --repo path: %w", err)
			}
			if err := protocol.WriteCurrentFeature(repoAbs, feature); err != nil {
				return fmt.Errorf("recording current feature: %w", err)
			}

			fmt.Printf("Local-mode coordination folder created at %s\n", localDir)
			fmt.Printf("%q recorded as the current feature for this worktree — every other `agentflow local` command here can now omit --feature.\n", feature)
			fmt.Println("Add it to .gitignore if it isn't already covered by `.agentflow/local/`.")
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	cmd.Flags().StringVar(&title, "title", "", "Human-readable feature title (default: feature ID)")
	cmd.Flags().StringVar(&root, "root", "develop", "Root branch")
	cmd.Flags().StringVar(&branch, "branch", "", "Feature branch")
	cmd.Flags().StringVar(&worktree, "worktree", "", "Worktree path")
	_ = cmd.MarkFlagRequired("feature")

	return cmd
}

func newLocalPromoteCmd() *cobra.Command {
	var repo, from, feature string

	cmd := &cobra.Command{
		Use:   "promote",
		Short: "Move a feature's pending coordination folder (created in the root repo, pre-worktree) into its worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoAbs, err := filepath.Abs(repo)
			if err != nil {
				return fmt.Errorf("invalid --repo path: %w", err)
			}
			if !gitops.IsGitRepo(repoAbs) {
				return fmt.Errorf("%s is not a git repository", repoAbs)
			}
			fromAbs, err := filepath.Abs(from)
			if err != nil {
				return fmt.Errorf("invalid --from path: %w", err)
			}
			if !gitops.IsGitRepo(fromAbs) {
				return fmt.Errorf("%s is not a git repository", fromAbs)
			}

			src := protocol.DefaultLocalPath(fromAbs, feature)
			if _, err := os.Stat(src); os.IsNotExist(err) {
				return fmt.Errorf("%s does not exist — run `agentflow local init` (and `agentflow local plan`) in the root repo first", src)
			}

			dst := protocol.DefaultLocalPath(repoAbs, feature)
			if _, err := os.Stat(dst); err == nil {
				return fmt.Errorf("%s already exists — a feature is promoted exactly once", dst)
			}

			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				return err
			}
			if err := moveDir(src, dst); err != nil {
				return fmt.Errorf("moving %s to %s: %w", src, dst, err)
			}

			if err := protocol.WriteCurrentFeature(repoAbs, feature); err != nil {
				return fmt.Errorf("recording current feature: %w", err)
			}

			if current, err := protocol.ReadCurrentFeature(fromAbs); err == nil && current == feature {
				_ = os.Remove(protocol.CurrentFeaturePath(fromAbs))
			}

			fmt.Printf("Promoted %s to %s\n", src, dst)
			fmt.Printf("%q recorded as the current feature for %s.\n", feature, repoAbs)
			fmt.Println("Now that branch/worktree/root are known, update POLICY.md's corresponding fields by hand.")
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the worktree the feature is being promoted into")
	cmd.Flags().StringVar(&from, "from", "", "Path to the root repo holding the pending coordination folder (required)")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("feature")

	return cmd
}

// moveDir relocates src to dst, trying a plain rename first (fast, atomic)
// and falling back to a recursive copy + delete when src/dst live on
// different filesystems (e.g. a worktree created outside the main repo's
// tree), which os.Rename cannot handle.
func moveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	if err := copyDir(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

// newLocalRoundFileCmd builds task/result/checkpoint — the three files
// overwritten each round, always archived to history/ first.
func newLocalRoundFileCmd(use, filename, short string) *cobra.Command {
	var repo, feature, file string

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir, err := resolveLocalDir(repo, feature)
			if err != nil {
				return err
			}
			if _, err := os.Stat(localDir); os.IsNotExist(err) {
				return fmt.Errorf("%s does not exist — run `agentflow local init` first", localDir)
			}

			content, err := readContent(file)
			if err != nil {
				return err
			}

			if err := archiveAndWrite(localDir, filename, content); err != nil {
				return err
			}

			fmt.Printf("Wrote %s\n", filepath.Join(localDir, filename))
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (default: this worktree's current feature, set by 'agentflow local init')")
	cmd.Flags().StringVar(&file, "file", "", "Read content from this file instead of stdin")

	return cmd
}

func newLocalPlanCmd() *cobra.Command {
	return newLocalRoundFileCmd("plan", "PLAN.md", "Write PLAN.md (problem analysis, impact assessment, and the sequenced work plan)")
}

func newLocalTaskCmd() *cobra.Command {
	return newLocalRoundFileCmd("task", "task.md", "Write task.md (Controller assigns work to the Coder)")
}

func newLocalResultCmd() *cobra.Command {
	return newLocalRoundFileCmd("result", "result.md", "Write result.md (Coder reports back to the Controller)")
}

func newLocalCheckpointCmd() *cobra.Command {
	return newLocalRoundFileCmd("checkpoint", "checkpoint.md", "Overwrite checkpoint.md at a milestone")
}

func newLocalDiscussCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discuss",
		Short: "Start or close an isolated human deliberation",
	}
	cmd.AddCommand(newLocalDiscussStartCmd())
	cmd.AddCommand(newLocalDiscussCloseCmd())
	return cmd
}

func newLocalDiscussStartCmd() *cobra.Command {
	var repo, feature string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Allocate the next discussion ID and its folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir, err := resolveLocalDir(repo, feature)
			if err != nil {
				return err
			}
			discussionsDir := filepath.Join(localDir, "discussions")
			if _, err := os.Stat(discussionsDir); os.IsNotExist(err) {
				return fmt.Errorf("%s does not exist — run `agentflow local init` first", localDir)
			}

			id, err := nextDiscussionID(discussionsDir)
			if err != nil {
				return err
			}
			dir := filepath.Join(discussionsDir, id)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			fmt.Printf("Discussion %s started at %s\n", id, dir)
			fmt.Println("Deliberate with the Human, then close with:")
			if feature != "" {
				fmt.Printf("  agentflow local discuss close --repo %s --feature %s --id %s\n", repo, feature, id)
			} else {
				fmt.Printf("  agentflow local discuss close --repo %s --id %s\n", repo, id)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (default: this worktree's current feature, set by 'agentflow local init')")

	return cmd
}

func newLocalDiscussCloseCmd() *cobra.Command {
	var repo, feature, id, file string

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Write discussions/<id>/OUTCOME.md, closing the deliberation",
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir, err := resolveLocalDir(repo, feature)
			if err != nil {
				return err
			}
			dir := filepath.Join(localDir, "discussions", id)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("%s does not exist — run `agentflow local discuss start` first", dir)
			}

			outcome := filepath.Join(dir, "OUTCOME.md")
			if _, err := os.Stat(outcome); err == nil {
				return fmt.Errorf("%s already exists — a discussion is closed exactly once", outcome)
			}

			content, err := readContent(file)
			if err != nil {
				return err
			}
			if err := os.WriteFile(outcome, []byte(content), 0644); err != nil {
				return err
			}

			fmt.Printf("Discussion %s closed: %s\n", id, outcome)
			fmt.Println("This is a Controller session-restart trigger — start a fresh session next.")
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (default: this worktree's current feature, set by 'agentflow local init')")
	cmd.Flags().StringVar(&id, "id", "", "Discussion ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "Read OUTCOME.md content from this file instead of stdin")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// nextDiscussionID scans discussionsDir for numeric subfolders and returns
// the next one, zero-padded to 3 digits (e.g. "002").
func nextDiscussionID(discussionsDir string) (string, error) {
	entries, err := os.ReadDir(discussionsDir)
	if err != nil {
		return "", err
	}

	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if n, err := strconv.Atoi(e.Name()); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%03d", max+1), nil
}

func newLocalStatusCmd() *cobra.Command {
	var repo, feature string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print a read-only summary of the current local-mode state (for the Human)",
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir, err := resolveLocalDir(repo, feature)
			if err != nil {
				return err
			}
			result := protocol.ValidateLocal(localDir)

			fmt.Printf("Local folder: %s\n", localDir)
			if !result.Valid {
				fmt.Printf("Missing: %s\n", strings.Join(result.Missing, ", "))
			}

			for _, name := range []string{"POLICY.md", "PLAN.md", "checkpoint.md", "task.md", "result.md"} {
				path := filepath.Join(localDir, name)
				info, err := os.Stat(path)
				if err != nil {
					fmt.Printf("  %-14s (not written yet)\n", name)
					continue
				}
				line := fmt.Sprintf("  %-14s %d bytes, last modified %s", name, info.Size(), info.ModTime().UTC().Format(time.RFC3339))
				if name == "PLAN.md" {
					if data, err := os.ReadFile(path); err == nil {
						done, total := planProgress(string(data))
						line += fmt.Sprintf(" (%d/%d done)", done, total)
					}
				}
				fmt.Println(line)
			}

			discussionsDir := filepath.Join(localDir, "discussions")
			entries, _ := os.ReadDir(discussionsDir)
			var ids []string
			for _, e := range entries {
				if e.IsDir() {
					ids = append(ids, e.Name())
				}
			}
			sort.Strings(ids)
			fmt.Printf("  discussions/   %d (%s)\n", len(ids), strings.Join(ids, ", "))

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (default: this worktree's current feature, set by 'agentflow local init')")

	return cmd
}

// planProgress counts GFM task-list checkboxes ("- [ ]" / "- [x]") in a
// PLAN.md, giving `agentflow local status` a cheap way to show progress
// through the plan without parsing it as markdown.
func planProgress(content string) (done, total int) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- [x]"), strings.HasPrefix(trimmed, "- [X]"):
			done++
			total++
		case strings.HasPrefix(trimmed, "- [ ]"):
			total++
		}
	}
	return done, total
}

func newLocalContextCmd() *cobra.Command {
	var repo, feature, role string

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Print exactly the files a fresh Controller or Coder session should load",
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir, err := resolveLocalDir(repo, feature)
			if err != nil {
				return err
			}

			var files []string
			switch role {
			case "coder":
				files = []string{"POLICY.md", "checkpoint.md", "task.md"}
			case "controller":
				files = []string{"POLICY.md", "PLAN.md", "checkpoint.md", "task.md", "result.md"}
			default:
				return fmt.Errorf("--role must be %q or %q", "controller", "coder")
			}

			for _, name := range files {
				path := filepath.Join(localDir, name)
				data, err := os.ReadFile(path)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return err
				}
				fmt.Printf("--- %s ---\n%s\n", name, string(data))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (default: this worktree's current feature, set by 'agentflow local init')")
	cmd.Flags().StringVar(&role, "role", "", "controller or coder (required)")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}
