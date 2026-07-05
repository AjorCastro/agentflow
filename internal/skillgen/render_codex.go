package skillgen

import (
	"fmt"
	"strings"
)

// renderCodex reproduces Codex CLI's skill convention: YAML frontmatter with
// a verbose trigger-phrase description, a numbered "Core Workflow" overview,
// unnumbered "### Title" subsections under "Step Detail", a "Constraints"
// section, a "Success Criteria" section, and a "Next Step" blockquote.
func renderCodex(def SkillDef) string {
	var b strings.Builder

	fmt.Fprintf(&b, "---\nname: %s\ndescription: %s\n---\n\n", def.Name, def.DescriptionLong)
	fmt.Fprintf(&b, "# %s\n\n", def.Title)

	b.WriteString("## Core Workflow\n\n")
	for i, s := range def.Steps {
		fmt.Fprintf(&b, "%d. %s\n", i+1, s.Title)
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "## Idempotency\n\n%s\n\n", def.Idempotency)

	b.WriteString("## Step Detail\n\n")
	for _, s := range def.Steps {
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", s.Title, s.Body)
	}

	if len(def.Constraints) > 0 {
		b.WriteString("## Constraints\n\n")
		for _, c := range def.Constraints {
			fmt.Fprintf(&b, "- %s\n", c)
		}
		b.WriteString("\n")
	}

	if len(def.SuccessCriteria) > 0 {
		b.WriteString("## Success Criteria\n\n")
		for _, c := range def.SuccessCriteria {
			fmt.Fprintf(&b, "- %s\n", c)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Next Step\n\n")
	fmt.Fprintf(&b, "Tell the user:\n> %s\n\n", def.NextStepMessage)
	b.WriteString("Stop.\n")

	return b.String()
}
