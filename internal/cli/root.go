package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "agentflow",
		Short: "AgentFlow — repo-centric coordination workspace for human + AI agents",
		Long: `AgentFlow initializes a structured exchange folder in a Git repository
to coordinate work between a Human, a Web Reviewer, and a CLI Coding Agent.`,
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newValidateCmd())

	return root
}
