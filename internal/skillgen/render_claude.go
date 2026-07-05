package skillgen

import (
	"fmt"
	"strings"
)

// renderClaude reproduces Claude Code's skill convention: a plain "# Title"
// heading (no YAML frontmatter), a single unnumbered "## Steps" section
// containing numbered "### N. Title" subsections, an optional literal
// "Print summary" fenced block, and a final numbered "Stop" step.
func renderClaude(def SkillDef) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", def.Title)
	if def.Intro != "" {
		fmt.Fprintf(&b, "%s\n\n", def.Intro)
	}
	fmt.Fprintf(&b, "## Idempotency\n\n%s\n\n", def.Idempotency)
	b.WriteString("## Steps\n\n")

	n := 1
	for _, s := range def.Steps {
		fmt.Fprintf(&b, "### %d. %s\n\n%s\n", n, s.Title, s.Body)
		if s.ClaudeExtra != "" {
			fmt.Fprintf(&b, "\n%s\n", s.ClaudeExtra)
		}
		b.WriteString("\n")
		n++
	}

	if def.SummaryTemplate != "" {
		fmt.Fprintf(&b, "### %d. Print summary\n\n```\n%s\n```\n\n", n, def.SummaryTemplate)
		n++
	} else if def.NextStepMessage != "" {
		fmt.Fprintf(&b, "### %d. Tell the Human\n\n> %s\n\n", n, def.NextStepMessage)
		n++
	}

	fmt.Fprintf(&b, "### %d. Stop\n\n%s\n", n, def.StopMessage)

	return b.String()
}
