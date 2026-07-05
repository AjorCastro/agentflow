package cli

import (
	"fmt"
	"path/filepath"

	"github.com/AjorCastro/agentflow/internal/gitops"
	"github.com/spf13/cobra"
)

type docsSyncOptions struct {
	repo string
	root string
	push bool
}

func newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage AgentFlow-generated documentation",
	}
	cmd.AddCommand(newDocsSyncCmd())
	return cmd
}

func newDocsSyncCmd() *cobra.Command {
	opts := &docsSyncOptions{}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Write or refresh docs/WEB-AGENT-ROLE.md from the binary's canonical content",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsSync(opts)
		},
	}

	cmd.Flags().StringVar(&opts.repo, "repo", ".", "Path to the git repository")
	cmd.Flags().StringVar(&opts.root, "root", "develop", "Root branch docs/WEB-AGENT-ROLE.md lives on")
	cmd.Flags().BoolVar(&opts.push, "push", true, "Push if a commit was made (default true; use --push=false to skip)")

	return cmd
}

func runDocsSync(opts *docsSyncOptions) error {
	repoAbs, err := filepath.Abs(opts.repo)
	if err != nil {
		return fmt.Errorf("invalid --repo path: %w", err)
	}

	if !gitops.IsGitRepo(repoAbs) {
		return fmt.Errorf("%s is not a git repository", repoAbs)
	}

	if err := refreshWebAgentRoleDoc(repoAbs, opts.root, opts.push); err != nil {
		return err
	}

	fmt.Println("docs/WEB-AGENT-ROLE.md is up to date.")
	return nil
}
