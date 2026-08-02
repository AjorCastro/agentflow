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
     session, invoke the agentflow-local-init skill. It creates the worktree
     if needed, runs 'agentflow local init', drafts POLICY.md with you, and
     writes the first task.md:
       agentflow local init --feature <name> --branch feature/<name> --worktree .worktrees/<name>

  2. In a second, independent session opened in the same worktree, invoke
     the agentflow-local-coder-init skill. It confirms the link and loads
     only POLICY.md+checkpoint.md+task.md via:
       agentflow local context --role coder
     then reports back with:
       agentflow local result <<'EOF' ... EOF

  3. Back in the Controller session, review result.md, assign the next task
     or close the feature. Check progress any time with:
       agentflow local status

  4. For every session after the first one, use agentflow-local-resume
     (Controller) or agentflow-local-coder-resume (Coder) instead of the
     -init skills — including after a forced session restart mid-feature.

  5. If the Human needs to interrupt a session off a natural milestone (end
     of day, an unrelated interruption), use the agentflow-local-pause skill
     first so the next resume doesn't start blind.

  6. When the feature is merged, close it exactly like GitHub mode:
       agentflow close --branch feature/<name> --root develop
     This also removes .agentflow/local/<name>/ — nothing to clean up by hand.`,
	}

	cmd.AddCommand(newLocalInitCmd())
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

			for _, name := range []string{"POLICY.md", "checkpoint.md", "task.md", "result.md"} {
				path := filepath.Join(localDir, name)
				info, err := os.Stat(path)
				if err != nil {
					fmt.Printf("  %-14s (not written yet)\n", name)
					continue
				}
				fmt.Printf("  %-14s %d bytes, last modified %s\n", name, info.Size(), info.ModTime().UTC().Format(time.RFC3339))
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
				files = []string{"POLICY.md", "checkpoint.md", "task.md", "result.md"}
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
