package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AjorCastro/agentflow/internal/gitops"
	"github.com/AjorCastro/agentflow/internal/protocol"
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
	cmd.Flags().BoolVar(&opts.push, "push", true, "Push branch to origin after init (default true; use --push=false to skip)")
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

	// Validate remote exists if push is requested.
	if opts.push && !gitops.HasRemote(repoAbs) {
		return fmt.Errorf(
			"--push requires a remote configured in this repository\n" +
			"Add one first, e.g.:\n" +
			"  git remote add origin git@github.com:<user>/<repo>.git",
		)
	}

	// Keep docs/WEB-AGENT-ROLE.md on the root branch in sync with the binary's
	// canonical content, so projects set up with an older agentflow version
	// don't drift out of date. Every init is a natural sync point.
	if err := refreshWebAgentRoleDoc(repoAbs, opts.root, opts.push); err != nil {
		return err
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

// refreshWebAgentRoleDoc overwrites docs/WEB-AGENT-ROLE.md on the root branch
// with the current canonical content, committing and pushing only if it
// actually changed. Skipped (non-fatal) if repoAbs isn't currently checked
// out on root — we never implicitly switch branches for the user.
func refreshWebAgentRoleDoc(repoAbs, root string, push bool) error {
	current, err := gitops.CurrentBranch(repoAbs)
	if err != nil {
		return err
	}
	if current != root {
		fmt.Printf("Note: %s is on branch %q, not %q — skipping docs/WEB-AGENT-ROLE.md refresh.\n", repoAbs, current, root)
		return nil
	}

	docRel := filepath.Join("docs", "WEB-AGENT-ROLE.md")
	docAbs := filepath.Join(repoAbs, docRel)
	content := protocol.WebAgentRoleMD()

	if existing, err := os.ReadFile(docAbs); err == nil && string(existing) == content {
		return nil
	}

	if err := os.MkdirAll(filepath.Join(repoAbs, "docs"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(docAbs, []byte(content), 0644); err != nil {
		return err
	}

	changed, err := gitops.PathHasChanges(repoAbs, docRel)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	fmt.Println("Refreshing docs/WEB-AGENT-ROLE.md with the latest canonical content...")
	if err := gitops.AddPathAndCommit(repoAbs, docRel, "chore: refresh docs/WEB-AGENT-ROLE.md"); err != nil {
		return err
	}
	if push && gitops.HasRemote(repoAbs) {
		if err := gitops.PushBranch(repoAbs, root); err != nil {
			return err
		}
	}
	return nil
}

func createExchangeStructure(exchangeDir, featureID, title, root, branch, worktree, exchangeRel string) error {
	// Create subdirectories.
	for _, sub := range protocol.RequiredDirs {
		if err := os.MkdirAll(filepath.Join(exchangeDir, sub), 0755); err != nil {
			return err
		}
	}

	// Write root files.
	files := map[string]string{
		"README.md": protocol.ReadmeMD(featureID, title),
		"STATUS.md": protocol.StatusMD(),
		"CONFIG.md": protocol.ConfigMD(featureID, title, root, branch, worktree, exchangeRel),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(exchangeDir, name), []byte(content), 0644); err != nil {
			return err
		}
	}

	// Write prompt files.
	prompts := map[string]string{
		"web-agent-role.md": protocol.WebAgentRoleMD(),
	}
	for name, content := range prompts {
		if err := os.WriteFile(filepath.Join(exchangeDir, "prompts", name), []byte(content), 0644); err != nil {
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
