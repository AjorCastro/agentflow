package cli

import (
	"fmt"
	"path/filepath"

	"github.com/agentflow/agentflow/internal/gitops"
	"github.com/agentflow/agentflow/internal/protocol"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show AgentFlow workspace status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(repo)
		},
	}

	cmd.Flags().StringVar(&repo, "repo", ".", "Path to the git repository")
	return cmd
}

func runStatus(repo string) error {
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return fmt.Errorf("invalid --repo path: %w", err)
	}

	if !gitops.IsGitRepo(repoAbs) {
		return fmt.Errorf("%s is not a git repository", repoAbs)
	}

	branch, err := gitops.CurrentBranch(repoAbs)
	if err != nil {
		return err
	}

	gitStatus, err := gitops.ShortStatus(repoAbs)
	if err != nil {
		return err
	}

	fmt.Printf("Repo    : %s\n", repoAbs)
	fmt.Printf("Branch  : %s\n", branch)

	if gitStatus == "" {
		fmt.Println("Git     : clean")
	} else {
		fmt.Printf("Git     :\n%s\n", gitStatus)
	}

	// Find AgentFlow exchange folders.
	exchanges, err := protocol.FindExchangeFolders(repoAbs)
	if err != nil {
		return err
	}

	if len(exchanges) == 0 {
		fmt.Println()
		fmt.Println("No AgentFlow workspace found in this repository.")
		fmt.Println("Run 'agentflow init --branch <branch> --worktree <path>' to create one.")
		return nil
	}

	for _, ex := range exchanges {
		statePath := filepath.Join(ex, "state.json")
		state, err := protocol.ReadStateFile(statePath)
		if err != nil {
			fmt.Printf("\nExchange: %s\n  (could not read state.json: %v)\n", ex, err)
			continue
		}

		fmt.Println()
		fmt.Printf("Exchange      : %s\n", ex)
		fmt.Printf("  feature_id  : %s\n", state.FeatureID)
		fmt.Printf("  title       : %s\n", state.Title)
		fmt.Printf("  phase       : %s\n", state.CurrentPhase)
		fmt.Printf("  turn        : %s\n", state.CurrentTurn)
		fmt.Printf("  status      : %s\n", state.Status)
		fmt.Printf("  next_action : %s\n", state.NextAction)
		if state.AttentionRequired {
			fmt.Println("  ⚠  attention required")
		}
	}

	return nil
}
