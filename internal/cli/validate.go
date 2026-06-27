package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/agentflow/agentflow/internal/gitops"
	"github.com/agentflow/agentflow/internal/protocol"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate AgentFlow workspace structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(repo)
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	return cmd
}

func runValidate(repo string) error {
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return fmt.Errorf("invalid --repo path: %w", err)
	}

	if !gitops.IsGitRepo(repoAbs) {
		return fmt.Errorf("%s is not a git repository", repoAbs)
	}

	exchanges, err := protocol.FindExchangeFolders(repoAbs)
	if err != nil {
		return err
	}

	if len(exchanges) == 0 {
		fmt.Println("No AgentFlow workspace found in this repository.")
		fmt.Println("Run 'agentflow init --branch <branch> --worktree <path>' to create one.")
		return nil
	}

	allValid := true
	for _, ex := range exchanges {
		result := protocol.ValidateExchange(ex)
		rel, _ := filepath.Rel(repoAbs, ex)

		if result.Valid {
			fmt.Printf("✓ %s — valid\n", rel)
		} else {
			allValid = false
			fmt.Printf("✗ %s — missing: %s\n", rel, strings.Join(result.Missing, ", "))
		}
	}

	if !allValid {
		return fmt.Errorf("one or more AgentFlow workspaces are incomplete")
	}
	return nil
}
