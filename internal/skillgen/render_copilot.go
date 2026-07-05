package skillgen

import (
	"fmt"
	"strings"
)

// renderCopilot reproduces Copilot CLI's skill convention: YAML frontmatter
// with a terse one-sentence description, no "Core Workflow" overview and no
// "Constraints" section (those facts are folded inline into the relevant
// step body instead), numbered "### N. Title" subsections under
// "## Procedure", a "Success criteria" section, and a "Next step" blockquote.
func renderCopilot(def SkillDef) string {
	var b strings.Builder

	fmt.Fprintf(&b, "---\nname: %s\ndescription: %s\n---\n\n", def.Name, def.DescriptionShort)

	b.WriteString("## Objective\n\n")
	if def.Intro != "" {
		fmt.Fprintf(&b, "%s\n\n", def.Intro)
	}

	fmt.Fprintf(&b, "## Idempotency\n\n%s\n\n", def.Idempotency)

	steps := def.Steps
	if def.HasSeparatePrereqSection && len(steps) > 0 {
		fmt.Fprintf(&b, "## Prerequisites\n\n%s\n\n", steps[0].Body)
		steps = steps[1:]
	}

	b.WriteString("## Procedure\n\n")
	for i, s := range steps {
		fmt.Fprintf(&b, "### %d. %s\n\n%s\n\n", i+1, s.Title, s.Body)
	}

	if len(def.SuccessCriteria) > 0 {
		b.WriteString("## Success criteria\n\n")
		for _, c := range def.SuccessCriteria {
			fmt.Fprintf(&b, "- %s\n", c)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Next step\n\n")
	fmt.Fprintf(&b, "Tell the user:\n> %s\n\n", def.NextStepMessage)
	b.WriteString("Stop.\n")

	return b.String()
}
