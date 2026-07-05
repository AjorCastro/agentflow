package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AjorCastro/agentflow/internal/gitops"
	"github.com/AjorCastro/agentflow/internal/protocol"
	"github.com/spf13/cobra"
)

type closeOptions struct {
	repo         string
	root         string
	branch       string
	deleteRemote bool
	force        bool
}

func newCloseCmd() *cobra.Command {
	opts := &closeOptions{}

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close a feature: remove worktree and delete branch after merge",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClose(opts)
		},
	}

	cmd.Flags().StringVar(&opts.repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&opts.root, "root", "develop", "Root branch the feature was merged into")
	cmd.Flags().StringVar(&opts.branch, "branch", "", "Feature branch to close (required)")
	cmd.Flags().BoolVar(&opts.deleteRemote, "delete-remote", false, "Also delete the remote branch")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Force removal even if branch is not fully merged")

	_ = cmd.MarkFlagRequired("branch")

	return cmd
}

func runClose(opts *closeOptions) error {
	repoAbs, err := filepath.Abs(opts.repo)
	if err != nil {
		return fmt.Errorf("invalid --repo path: %w", err)
	}

	if !gitops.IsGitRepo(repoAbs) {
		return fmt.Errorf("%s is not a git repository", repoAbs)
	}

	// Check branch exists.
	if !gitops.BranchExists(repoAbs, opts.branch) {
		return fmt.Errorf("branch %q does not exist", opts.branch)
	}

	// Check branch is merged (unless --force).
	if !opts.force {
		if !gitops.IsBranchMerged(repoAbs, opts.branch) {
			return fmt.Errorf(
				"branch %q does not appear to be merged\nUse --force to close anyway",
				opts.branch,
			)
		}
	}

	// Remove linked worktree for this branch, if any.
	worktreePath, err := gitops.FindWorktreeForBranch(repoAbs, opts.branch)
	if err != nil {
		return err
	}
	if worktreePath != "" {
		fmt.Printf("Removing worktree at %s ...\n", worktreePath)
		if err := gitops.RemoveWorktree(repoAbs, worktreePath, opts.force); err != nil {
			return err
		}
	} else {
		fmt.Println("No linked worktree found for this branch.")
	}

	// Delete local branch.
	fmt.Printf("Deleting local branch %s ...\n", opts.branch)
	if err := gitops.DeleteBranch(repoAbs, opts.branch, opts.force); err != nil {
		return err
	}

	// Optionally delete remote branch.
	if opts.deleteRemote {
		if !gitops.HasRemote(repoAbs) {
			fmt.Println("No remote configured, skipping remote branch deletion.")
		} else {
			fmt.Printf("Deleting remote branch %s ...\n", opts.branch)
			if err := gitops.DeleteRemoteBranch(repoAbs, opts.branch); err != nil {
				return err
			}
		}
	}

	// Remove the merged exchange folder from the root branch, if present.
	// The feature branch carries .agentflow/features/<feature-id>/ as regular
	// tracked files, so merging it into root leaves that folder behind —
	// deleting the worktree and the branch alone does not clean it up.
	featureID := protocol.FeatureIDFromBranch(opts.branch)
	exchangeRel := filepath.Join(".agentflow", "features", featureID)
	exchangeAbs := filepath.Join(repoAbs, exchangeRel)

	if _, err := os.Stat(exchangeAbs); err == nil {
		currentBranch, err := gitops.CurrentBranch(repoAbs)
		if err != nil {
			return err
		}
		if currentBranch != opts.root {
			fmt.Printf(
				"\n%s still exists but %s is on branch %q, not %q — skipping cleanup.\n"+
					"Run 'agentflow close --branch %s --root %s' again from %q to remove it.\n",
				exchangeRel, repoAbs, currentBranch, opts.root, opts.branch, opts.root, opts.root,
			)
		} else {
			fmt.Printf("Removing merged exchange folder %s from %s ...\n", exchangeRel, opts.root)
			msg := fmt.Sprintf("chore: remove exchange folder for closed feature %s", opts.branch)
			if err := gitops.RemovePathAndCommit(repoAbs, exchangeRel, msg); err != nil {
				return err
			}
			if gitops.HasRemote(repoAbs) {
				if err := gitops.PushBranch(repoAbs, opts.root); err != nil {
					return err
				}
			}
		}
	}

	fmt.Println()
	fmt.Printf("Feature %q closed.\n", opts.branch)
	fmt.Println()
	fmt.Println("Next action:")
	fmt.Println("  Run 'agentflow init --branch <new-branch> --worktree <path>' to start a new feature.")

	return nil
}
