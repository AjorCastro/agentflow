package skillgen

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var allSkillsForTest = []SkillDef{InitSkill, SetupSkill, TurnSkill, CloseSkill, LocalControllerSkill, LocalCoderSkill}

// TestGeneratedFilesAreFresh fails if the checked-in skill files in skills/
// don't match what Render currently produces — i.e. someone hand-edited a
// .md/SKILL.md file without updating its SkillDef, or changed a SkillDef
// without running `go generate ./...`. This is the safety net the whole
// skillgen package exists to provide.
func TestGeneratedFilesAreFresh(t *testing.T) {
	repoRoot := repoRootDir(t)

	cases := []struct {
		flavor Flavor
		path   func(name string) string
	}{
		{FlavorClaude, func(name string) string { return filepath.Join(repoRoot, "skills", name+".md") }},
		{FlavorCodex, func(name string) string { return filepath.Join(repoRoot, "skills", "codex", name, "SKILL.md") }},
		{FlavorCopilot, func(name string) string { return filepath.Join(repoRoot, "skills", "copilot", name, "SKILL.md") }},
	}

	for _, def := range allSkillsForTest {
		for _, c := range cases {
			path := c.path(def.Name)
			want := Render(def, c.flavor)

			got, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s: %v (run `go generate ./...` from the repo root)", path, err)
				continue
			}
			if string(got) != want {
				t.Errorf("%s is stale — it doesn't match Render(%s, flavor). Run `go generate ./...` from the repo root and commit the result.", path, def.Name)
			}
		}
	}
}

// repoRootDir finds the module root by walking up from this test file's
// directory until a go.mod is found.
func repoRootDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

func TestRenderClaude_NoFrontmatter(t *testing.T) {
	out := Render(InitSkill, FlavorClaude)
	if strings.HasPrefix(out, "---") {
		t.Error("Claude skill file should not have YAML frontmatter")
	}
	if !strings.HasPrefix(out, "# "+InitSkill.Title) {
		t.Error("Claude skill file should start with a plain H1 title")
	}
}

func TestRenderCodexAndCopilot_HaveFrontmatter(t *testing.T) {
	for _, flavor := range []Flavor{FlavorCodex, FlavorCopilot} {
		out := Render(InitSkill, flavor)
		if !strings.HasPrefix(out, "---\nname: "+InitSkill.Name+"\n") {
			t.Errorf("flavor %v: expected YAML frontmatter starting with name field", flavor)
		}
	}
}

var stepHeadingRe = regexp.MustCompile(`(?m)^### (\d+)\.`)

func TestRenderClaudeAndCopilot_StepNumberingHasNoDuplicates(t *testing.T) {
	for _, def := range allSkillsForTest {
		for _, flavor := range []Flavor{FlavorClaude, FlavorCopilot} {
			out := Render(def, flavor)
			matches := stepHeadingRe.FindAllStringSubmatch(out, -1)
			seen := map[string]bool{}
			for _, m := range matches {
				n := m[1]
				if seen[n] {
					t.Errorf("%s/%v: duplicate numbered step heading %q", def.Name, flavor, n)
				}
				seen[n] = true
			}
		}
	}
}

func TestRenderTurn_AgentTokensSubstituted(t *testing.T) {
	for _, flavor := range []Flavor{FlavorClaude, FlavorCodex, FlavorCopilot} {
		out := Render(TurnSkill, flavor)
		if strings.Contains(out, "{{AGENT_NAME}}") || strings.Contains(out, "{{AGENT_SLUG}}") {
			t.Errorf("flavor %v: unresolved agent token in output", flavor)
		}
	}
	claude := Render(TurnSkill, FlavorClaude)
	if !strings.Contains(claude, "Claude Code") || !strings.Contains(claude, "claude-code") {
		t.Error("Claude output should mention its own agent name and slug")
	}
	if !strings.Contains(claude, ".claude/handoffs/LATEST.md") {
		t.Error("Claude output should keep its LATEST.md auto-injection note")
	}

	codex := Render(TurnSkill, FlavorCodex)
	if strings.Contains(codex, "LATEST.md") {
		t.Error("Codex output should not include Claude's LATEST.md note")
	}
}

func TestRenderClose_MentionsForceInAllFlavors(t *testing.T) {
	for _, flavor := range []Flavor{FlavorClaude, FlavorCodex, FlavorCopilot} {
		out := Render(CloseSkill, flavor)
		if !strings.Contains(out, "--force") {
			t.Errorf("flavor %v: agentflow-close should mention --force in every flavor (this was drift bug #2)", flavor)
		}
	}
}

func TestRenderInit_ChecksNextActionInAllFlavors(t *testing.T) {
	for _, flavor := range []Flavor{FlavorClaude, FlavorCodex, FlavorCopilot} {
		out := Render(InitSkill, flavor)
		if !strings.Contains(out, "next_action") {
			t.Errorf("flavor %v: agentflow-init should verify next_action in every flavor (this was drift bug #1)", flavor)
		}
	}
}

func TestRenderSetup_NeverEmbedsWebAgentRoleTextDirectly(t *testing.T) {
	for _, flavor := range []Flavor{FlavorClaude, FlavorCodex, FlavorCopilot} {
		out := Render(SetupSkill, flavor)
		if !strings.Contains(out, "agentflow docs sync") {
			t.Errorf("flavor %v: agentflow-setup should tell the agent to run `agentflow docs sync` instead of hand-writing docs/WEB-AGENT-ROLE.md", flavor)
		}
		if strings.Contains(out, "Web Reviewer** in the AgentFlow protocol") {
			t.Errorf("flavor %v: agentflow-setup should not hand-embed WEB-AGENT-ROLE.md content anymore (this was drift bug #4)", flavor)
		}
	}
}
