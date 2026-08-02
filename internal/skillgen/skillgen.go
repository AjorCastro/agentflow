// Package skillgen generates the AgentFlow CLI-agent skill files (Claude Code,
// Codex CLI, Copilot CLI) from a single Go data source per skill, instead of
// maintaining 3 hand-written copies of each skill's steps in sync by hand.
//
// Each skill is described once as a SkillDef (the facts: what to check, what
// commands to run, what to create). A per-flavor Render function turns that
// same SkillDef into the markdown structure that flavor's agent expects
// (YAML frontmatter or not, numbered steps or "Step Detail" subsections,
// "Success Criteria" section or a literal summary block, etc.).
package skillgen

import "strings"

// Flavor identifies which CLI agent a skill file is rendered for.
type Flavor int

const (
	FlavorClaude Flavor = iota
	FlavorCodex
	FlavorCopilot
	FlavorKimi
	FlavorOpenCode
)

// Step is one numbered (or sub-headed) unit of a skill's instructions.
type Step struct {
	// Title is the step's short name, e.g. "Pull latest changes from origin".
	Title string
	// Body is the step's markdown body (can span paragraphs, lists, fences).
	Body string
	// ClaudeExtra is an optional paragraph appended only when rendering for
	// Claude, for behavior that's genuinely specific to that agent (e.g.
	// Claude Code auto-injects .claude/handoffs/LATEST.md as session
	// context, which Codex/Copilot have no equivalent mechanism for).
	ClaudeExtra string
}

// SkillDef is the single source of truth for one skill (e.g. "agentflow-init"),
// rendered into all 3 flavors by Render.
type SkillDef struct {
	// Name is the skill's identifier, e.g. "agentflow-init". Used as the
	// Claude filename (<Name>.md) and the codex/copilot folder name.
	Name string
	// Title is the human-readable heading, e.g. "AgentFlow Init".
	Title string

	// DescriptionLong is the codex frontmatter description: verbose, with an
	// explicit enumeration of trigger phrases (codex matches skills by
	// keyword against this field).
	DescriptionLong string
	// DescriptionShort is the copilot frontmatter description: a single
	// terse sentence.
	DescriptionShort string

	// Intro is 1-2 sentences shown right after the title, before Idempotency.
	Intro string

	// Idempotency is this skill's check-before-acting paragraph. It differs
	// in wording between skills (agentflow-turn's idempotency check is about
	// not overwriting phase results; agentflow-close's is about not
	// re-removing an already-removed worktree) but is identical in meaning
	// across all 3 flavors of the same skill, so it lives here once instead
	// of 3 times.
	Idempotency string

	// HasSeparatePrereqSection is true only for agentflow-setup: Copilot's
	// convention breaks prerequisite checks out into their own top-level
	// "## Prerequisites" section instead of folding them into the first
	// numbered procedure step (which is what Claude and Codex do — for them
	// the prerequisites check is simply Steps[0]).
	HasSeparatePrereqSection bool

	// Steps are the ordered core instructions. Numbering is computed by the
	// renderer, never hand-written, so the kind of duplicate-numbering bug
	// found in the old hand-written Copilot agentflow-turn file can't recur.
	Steps []Step

	// Constraints are "never do X" rules. Rendered as a distinct
	// "## Constraints" section only for Codex (matching its existing
	// convention); Claude and Copilot fold the same facts inline into the
	// relevant step's body instead, so this list is still the single source
	// of truth for the fact even though it's structurally placed differently.
	Constraints []string

	// SuccessCriteria are checkable "this worked" bullets, rendered as a
	// "Success Criteria"/"Success criteria" section for Codex and Copilot.
	SuccessCriteria []string

	// SummaryTemplate is a literal fenced text block Claude prints as its
	// second-to-last step (e.g. "Branch : <branch>\nWorktree : <worktree>").
	// Empty if this skill has no such template (matches current convention:
	// only init/setup/close have one; turn does not).
	SummaryTemplate string

	// NextStepMessage is the blockquote message told to the Human once the
	// skill finishes (e.g. `"Done. The Web Reviewer can now review ..."`).
	NextStepMessage string

	// StopMessage is the final "do not continue" sentence(s).
	StopMessage string
}

// flavorSuffix returns the filename/handoff-agent-name suffix used by a
// flavor, e.g. "claude-code", "codex-cli", "copilot-cli".
func (f Flavor) agentName() string {
	switch f {
	case FlavorCodex:
		return "Codex CLI"
	case FlavorCopilot:
		return "Copilot CLI"
	case FlavorKimi:
		return "Kimi Code CLI"
	case FlavorOpenCode:
		return "OpenCode"
	default:
		return "Claude Code"
	}
}

func (f Flavor) handoffSlug() string {
	switch f {
	case FlavorCodex:
		return "codex-cli"
	case FlavorCopilot:
		return "copilot-cli"
	case FlavorKimi:
		return "kimi-cli"
	case FlavorOpenCode:
		return "opencode"
	default:
		return "claude-code"
	}
}

// Render dispatches to the flavor-specific renderer, then substitutes any
// {{AGENT_NAME}}/{{AGENT_SLUG}} tokens with this flavor's concrete values
// (e.g. a step describing a handoff file that must literally record which
// agent produced it).
func Render(def SkillDef, flavor Flavor) string {
	var out string
	switch flavor {
	case FlavorCodex, FlavorKimi, FlavorOpenCode:
		// Kimi Code CLI's and OpenCode's Agent Skills SKILL.md convention
		// (folder + SKILL.md, YAML frontmatter with name/description, name
		// limited to lowercase letters/numbers/hyphens matching the folder
		// name) is identical to Codex's, so they reuse the same renderer
		// rather than duplicating it.
		out = renderCodex(def)
	case FlavorCopilot:
		out = renderCopilot(def)
	default:
		out = renderClaude(def)
	}
	out = strings.ReplaceAll(out, "{{AGENT_NAME}}", flavor.agentName())
	out = strings.ReplaceAll(out, "{{AGENT_SLUG}}", flavor.handoffSlug())
	return out
}
