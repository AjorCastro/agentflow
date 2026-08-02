package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/AjorCastro/agentflow/skills"
	"github.com/spf13/cobra"
)

func newInstallSkillsCmd() *cobra.Command {
	var agent string

	cmd := &cobra.Command{
		Use:   "install-skills",
		Short: "Install AgentFlow skills for your CLI agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallSkills(agent)
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "claude", "Target agent: claude, copilot, codex, kimi, opencode")

	return cmd
}

func runInstallSkills(agent string) error {
	switch agent {
	case "claude":
		return installClaudeSkills()
	case "copilot":
		return installCopilotSkills()
	case "codex":
		return installCodexSkills()
	case "kimi":
		return installKimiSkills()
	case "opencode":
		return installOpenCodeSkills()
	default:
		return fmt.Errorf("unknown agent %q — supported: claude, copilot, codex, kimi, opencode", agent)
	}
}

// installClaudeSkills copies *.md files to ~/.claude/commands/
func installClaudeSkills() error {
	target := filepath.Join(os.Getenv("HOME"), ".claude", "commands")
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("could not create %s: %w", target, err)
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
			return fmt.Errorf("could not read %s: %w", entry.Name(), err)
		}
		dest := filepath.Join(target, entry.Name())
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", dest, err)
		}
		fmt.Printf("  installed: %s\n", dest)
		installed++
	}

	fmt.Printf("\n%d skill(s) installed in %s\n", installed, target)
	fmt.Println("\nAvailable in Claude Code:")
	fmt.Println("  /agentflow-setup   — set up a new project and push to GitHub")
	fmt.Println("  /agentflow-init    — initialize workspace from agentflow-init-*.md")
	fmt.Println("  /agentflow-turn    — execute the CLI Agent's turn")
	fmt.Println("  /agentflow-close   — clean up after merge")
	fmt.Println("  /agentflow-local-init         — bootstrap a new local-mode feature (Controller)")
	fmt.Println("  /agentflow-local-resume       — resume as Controller in local mode")
	fmt.Println("  /agentflow-local-coder-init   — link a new Coder session to a feature")
	fmt.Println("  /agentflow-local-coder-resume — resume as Coder in local mode")
	fmt.Println("  /agentflow-local-pause        — safely pause a Controller/Coder session mid-work")
	return nil
}

// installCodexSkills copies codex/<skill>/ folders to ~/.codex/skills/
func installCodexSkills() error {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		codexHome = filepath.Join(os.Getenv("HOME"), ".codex")
	}
	target := filepath.Join(codexHome, "skills")
	return installSkillFolders("codex", target, "Codex CLI", "")
}

// installCopilotSkills copies copilot/<skill>/ folders to ~/.copilot/skills/
func installCopilotSkills() error {
	target := filepath.Join(os.Getenv("HOME"), ".copilot", "skills")
	return installSkillFolders("copilot", target, "Copilot CLI", "run /skills reload to activate")
}

// installKimiSkills copies kimi/<skill>/ folders to <KIMI_CODE_HOME>/skills/
// (default ~/.kimi-code/skills/) — the user-level "brand" skills directory
// the installed kimi-code binary actually scans (confirmed by inspecting the
// binary: it resolves brandHomeDir from $KIMI_CODE_HOME, falling back to
// ~/.kimi-code, and looks for a "skills" subdirectory there — not ~/.kimi/).
func installKimiSkills() error {
	kimiHome := os.Getenv("KIMI_CODE_HOME")
	if kimiHome == "" {
		kimiHome = filepath.Join(os.Getenv("HOME"), ".kimi-code")
	}
	target := filepath.Join(kimiHome, "skills")
	return installSkillFolders("kimi", target, "Kimi Code CLI", "invoke with /skill:<name>, e.g. /skill:agentflow-turn")
}

// installOpenCodeSkills copies opencode/<skill>/ folders to
// ~/.config/opencode/skills/ — OpenCode's native global Agent Skills
// directory (no environment variable override is documented for it, unlike
// Codex's $CODEX_HOME or Kimi's $KIMI_CODE_HOME). Unlike the other agents'
// skills, OpenCode's Agent Skills are not user-typed slash commands: the
// model loads one on demand, guided by its `description` frontmatter.
func installOpenCodeSkills() error {
	target := filepath.Join(os.Getenv("HOME"), ".config", "opencode", "skills")
	return installSkillFolders("opencode", target, "OpenCode", "these are Agent Skills: the model loads them on demand, there is no /<name> to type")
}

// installSkillFolders copies embedded <srcDir>/<skill>/ folders to target.
func installSkillFolders(srcDir, target, agentLabel, hint string) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("could not create %s: %w", target, err)
	}

	entries, err := fs.ReadDir(skills.FS, srcDir)
	if err != nil {
		return fmt.Errorf("could not read embedded %s skills: %w", srcDir, err)
	}

	installed := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		destDir := filepath.Join(target, skillName)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %w", destDir, err)
		}

		skillFS, err := fs.Sub(skills.FS, filepath.Join(srcDir, skillName))
		if err != nil {
			return fmt.Errorf("could not access skill %s: %w", skillName, err)
		}
		skillFiles, _ := fs.ReadDir(skillFS, ".")
		for _, f := range skillFiles {
			if f.IsDir() {
				continue
			}
			data, err := fs.ReadFile(skillFS, f.Name())
			if err != nil {
				return fmt.Errorf("could not read %s/%s: %w", skillName, f.Name(), err)
			}
			dest := filepath.Join(destDir, f.Name())
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return fmt.Errorf("could not write %s: %w", dest, err)
			}
		}

		fmt.Printf("  installed: %s/\n", destDir)
		installed++
	}

	msg := fmt.Sprintf("\n%d skill(s) installed in %s\n", installed, target)
	if hint != "" {
		msg += fmt.Sprintf("(%s)\n", hint)
	}
	fmt.Print(msg)
	fmt.Printf("\nAvailable in %s:\n", agentLabel)
	fmt.Println("  agentflow-setup   — set up a new project and push to GitHub")
	fmt.Println("  agentflow-init    — initialize workspace from agentflow-init-*.md")
	fmt.Println("  agentflow-turn    — execute the CLI Agent's turn")
	fmt.Println("  agentflow-close   — clean up after merge")
	fmt.Println("  agentflow-local-init         — bootstrap a new local-mode feature (Controller)")
	fmt.Println("  agentflow-local-resume       — resume as Controller in local mode")
	fmt.Println("  agentflow-local-coder-init   — link a new Coder session to a feature")
	fmt.Println("  agentflow-local-coder-resume — resume as Coder in local mode")
	fmt.Println("  agentflow-local-pause        — safely pause a Controller/Coder session mid-work")
	return nil
}
