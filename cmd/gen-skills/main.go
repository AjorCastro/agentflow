// Command gen-skills writes the 36 AgentFlow skill files (9 skills × 4 agent
// flavors) from internal/skillgen's SkillDefs, into the exact paths that
// skills/embed.go already embeds. Run via `go generate ./...` from the repo
// root, or directly with `go run ./cmd/gen-skills [output-dir]`.
//
// It never changes skills/embed.go or internal/cli/install_skills.go — those
// keep embedding/copying whatever files exist at these paths, unaware that
// they're now generated rather than hand-written.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AjorCastro/agentflow/internal/skillgen"
)

var allSkills = []skillgen.SkillDef{
	skillgen.InitSkill,
	skillgen.SetupSkill,
	skillgen.TurnSkill,
	skillgen.CloseSkill,
	skillgen.LocalInitSkill,
	skillgen.LocalResumeSkill,
	skillgen.LocalCoderInitSkill,
	skillgen.LocalCoderResumeSkill,
	skillgen.LocalPauseSkill,
}

func main() {
	outDir := "."
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	for _, def := range allSkills {
		if err := writeFile(filepath.Join(outDir, def.Name+".md"), skillgen.Render(def, skillgen.FlavorClaude)); err != nil {
			fail(err)
		}
		if err := writeFile(filepath.Join(outDir, "codex", def.Name, "SKILL.md"), skillgen.Render(def, skillgen.FlavorCodex)); err != nil {
			fail(err)
		}
		if err := writeFile(filepath.Join(outDir, "copilot", def.Name, "SKILL.md"), skillgen.Render(def, skillgen.FlavorCopilot)); err != nil {
			fail(err)
		}
		if err := writeFile(filepath.Join(outDir, "kimi", def.Name, "SKILL.md"), skillgen.Render(def, skillgen.FlavorKimi)); err != nil {
			fail(err)
		}
	}

	fmt.Printf("Generated %d skill files (%d skills x 4 flavors) in %s\n", len(allSkills)*4, len(allSkills), outDir)
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	fmt.Println("  wrote", path)
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen-skills:", err)
	os.Exit(1)
}
