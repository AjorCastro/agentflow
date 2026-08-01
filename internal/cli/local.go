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

// resolveLocalDir returns the absolute local-mode coordination folder for
// --repo/--feature, validating repo is a git repository.
func resolveLocalDir(repo, feature string) (string, error) {
	if feature == "" {
		return "", fmt.Errorf("--feature is required")
	}
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return "", fmt.Errorf("invalid --repo path: %w", err)
	}
	if !gitops.IsGitRepo(repoAbs) {
		return "", fmt.Errorf("%s is not a git repository", repoAbs)
	}
	return protocol.DefaultLocalPath(repoAbs, feature), nil
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

			fmt.Printf("Local-mode coordination folder created at %s\n", localDir)
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
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "Read content from this file instead of stdin")
	_ = cmd.MarkFlagRequired("feature")

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
			fmt.Printf("  agentflow local discuss close --repo %s --feature %s --id %s\n", repo, feature, id)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	_ = cmd.MarkFlagRequired("feature")

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
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	cmd.Flags().StringVar(&id, "id", "", "Discussion ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "Read OUTCOME.md content from this file instead of stdin")
	_ = cmd.MarkFlagRequired("feature")
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
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	_ = cmd.MarkFlagRequired("feature")

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
	cmd.Flags().StringVar(&feature, "feature", "", "Feature ID (required)")
	cmd.Flags().StringVar(&role, "role", "", "controller or coder (required)")
	_ = cmd.MarkFlagRequired("feature")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}
