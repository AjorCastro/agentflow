package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflow/agentflow/internal/gitops"
	"github.com/agentflow/agentflow/internal/protocol"
	"github.com/spf13/cobra"
)

type initOptions struct {
	repo       string
	root       string
	branch     string
	worktree   string
	exchange   string
	title      string
	push       bool
	adopt      bool
	noWorktree bool
}

func newInitCmd() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize an AgentFlow workspace on a feature branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts)
		},
	}

	cmd.Flags().StringVar(&opts.repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&opts.root, "root", "develop", "Root branch to branch from")
	cmd.Flags().StringVar(&opts.branch, "branch", "", "Feature branch name (required)")
	cmd.Flags().StringVar(&opts.worktree, "worktree", "", "Worktree path (required unless --no-worktree)")
	cmd.Flags().StringVar(&opts.exchange, "exchange", "", "Exchange folder path (default: .agentflow/features/<feature-id>)")
	cmd.Flags().StringVar(&opts.title, "title", "", "Human-readable feature title")
	cmd.Flags().BoolVar(&opts.push, "push", false, "Push branch to origin after init")
	cmd.Flags().BoolVar(&opts.adopt, "adopt", false, "Adopt an existing branch instead of creating a new one")
	cmd.Flags().BoolVar(&opts.noWorktree, "no-worktree", false, "Skip worktree creation; operate on the current repo")

	_ = cmd.MarkFlagRequired("branch")

	return cmd
}

func runInit(opts *initOptions) error {
	// Resolve absolute repo path.
	repoAbs, err := filepath.Abs(opts.repo)
	if err != nil {
		return fmt.Errorf("invalid --repo path: %w", err)
	}

	// Validate git repo.
	if !gitops.IsGitRepo(repoAbs) {
		return fmt.Errorf("%s is not a git repository", repoAbs)
	}

	// Validate root branch.
	if !gitops.BranchExists(repoAbs, opts.root) {
		return fmt.Errorf("root branch %q does not exist in %s", opts.root, repoAbs)
	}

	// Derive feature ID early; exchange path is resolved after workDir is known.
	featureID := protocol.FeatureIDFromBranch(opts.branch)
	if opts.title == "" {
		opts.title = featureID
	}

	// Determine which directory the workspace will live in.
	workDir := repoAbs

	if !opts.noWorktree {
		if opts.worktree == "" {
			return fmt.Errorf("--worktree is required unless --no-worktree is set")
		}

		worktreeAbs, err := filepath.Abs(opts.worktree)
		if err != nil {
			return fmt.Errorf("invalid --worktree path: %w", err)
		}
		opts.worktree = worktreeAbs

		branchExists := gitops.BranchExists(repoAbs, opts.branch)
		if branchExists && !opts.adopt {
			return fmt.Errorf(
				"branch %q already exists; use --adopt to attach a worktree to it, or choose a different branch name",
				opts.branch,
			)
		}

		fmt.Printf("Creating worktree at %s ...\n", opts.worktree)
		if err := gitops.AddWorktree(repoAbs, opts.worktree, opts.branch, opts.root, branchExists && opts.adopt); err != nil {
			return err
		}
		workDir = opts.worktree
	} else {
		// No-worktree mode: must already be on the target branch.
		current, err := gitops.CurrentBranch(repoAbs)
		if err != nil {
			return err
		}
		if current != opts.branch {
			return fmt.Errorf(
				"--no-worktree requires the repo to be on branch %q, but HEAD is on %q",
				opts.branch, current,
			)
		}
		opts.worktree = "(none — operating on repo)"
	}

	// Default exchange path is relative to workDir (the worktree or repo).
	// Must be resolved after workDir is known.
	if opts.exchange == "" {
		opts.exchange = protocol.DefaultExchangePath(workDir, opts.branch)
	}

	// Resolve exchange path: relative paths are relative to workDir.
	exchangeAbs := opts.exchange
	if !filepath.IsAbs(exchangeAbs) {
		exchangeAbs = filepath.Join(workDir, opts.exchange)
	}

	// Create exchange folder structure.
	fmt.Printf("Creating exchange folder at %s ...\n", exchangeAbs)
	if err := createExchangeStructure(exchangeAbs, featureID, opts.title, opts.root, opts.branch, opts.worktree, opts.exchange); err != nil {
		return fmt.Errorf("failed to create exchange structure: %w", err)
	}

	// Initial commit.
	fmt.Println("Committing initial AgentFlow workspace...")
	if err := gitops.CommitAll(workDir, "Initialize AgentFlow workspace"); err != nil {
		return err
	}

	// Optional push.
	if opts.push {
		fmt.Printf("Pushing %s to origin...\n", opts.branch)
		if err := gitops.PushBranch(workDir, opts.branch); err != nil {
			return err
		}
	}

	// Summary.
	printInitSummary(repoAbs, opts, exchangeAbs)
	return nil
}

func createExchangeStructure(exchangeDir, featureID, title, root, branch, worktree, exchangeRel string) error {
	// Create subdirectories.
	for _, sub := range protocol.RequiredDirs {
		if err := os.MkdirAll(filepath.Join(exchangeDir, sub), 0755); err != nil {
			return err
		}
	}

	// Write files.
	files := map[string]string{
		"README.md":  protocol.ReadmeMD(featureID, title),
		"STATUS.md":  protocol.StatusMD(),
		"CONFIG.md":  protocol.ConfigMD(featureID, title, root, branch, worktree, exchangeRel),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(exchangeDir, name), []byte(content), 0644); err != nil {
			return err
		}
	}

	// Write state.json.
	state := protocol.NewInitialState(featureID, title, root, branch, worktree, exchangeRel)
	return protocol.WriteStateFile(filepath.Join(exchangeDir, "state.json"), state)
}

func printInitSummary(repo string, opts *initOptions, exchangeAbs string) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║         AgentFlow workspace initialized          ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("  Repo          : %s\n", repo)
	fmt.Printf("  Root branch   : %s\n", opts.root)
	fmt.Printf("  Feature branch: %s\n", opts.branch)
	fmt.Printf("  Worktree      : %s\n", opts.worktree)
	fmt.Printf("  Exchange      : %s\n", exchangeAbs)
	fmt.Println()
	fmt.Println("Next action:")
	fmt.Printf("  Open the worktree in your CLI agent and run the initial AgentFlow skill.\n")
	fmt.Printf("  Exchange folder is ready at: %s\n", exchangeAbs)
	fmt.Println()
}
