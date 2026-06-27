package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/agentflow/agentflow/skills"
	"github.com/spf13/cobra"
)

func newInstallSkillsCmd() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "install-skills",
		Short: "Install AgentFlow CLI skills into ~/.claude/commands/",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallSkills(target)
		},
	}

	defaultTarget := filepath.Join(os.Getenv("HOME"), ".claude", "commands")
	cmd.Flags().StringVar(&target, "target", defaultTarget, "Directory to install skills into")

	return cmd
}

func runInstallSkills(target string) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("could not create target directory %s: %w", target, err)
	}

	entries, err := fs.ReadDir(skills.FS, ".")
	if err != nil {
		return fmt.Errorf("could not read embedded skills: %w", err)
	}

	installed := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		data, err := skills.FS.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("could not read skill %s: %w", entry.Name(), err)
		}

		dest := filepath.Join(target, entry.Name())
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", dest, err)
		}

		fmt.Printf("  installed: %s\n", dest)
		installed++
	}

	fmt.Printf("\n%d skill(s) installed in %s\n", installed, target)
	fmt.Println("\nAvailable skills in Claude Code:")
	fmt.Println("  /agentflow-setup   — set up a new project and push to GitHub")
	fmt.Println("  /agentflow-init    — initialize workspace from agentflow-init.md")

	return nil
}
