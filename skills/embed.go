// Package skills embeds the AgentFlow CLI-agent skill files. The files
// themselves are generated from internal/skillgen — see
// cmd/gen-skills/main.go and the go:generate directive below. Never hand-edit
// the *.md/SKILL.md files directly; edit the SkillDef in internal/skillgen
// and regenerate instead.
package skills

import "embed"

//go:generate go run ../cmd/gen-skills

//go:embed *.md copilot codex
var FS embed.FS
