package skillgen

// TurnSkill is the single source of truth for the agentflow-turn skill
// across all 3 agent flavors.
var TurnSkill = SkillDef{
	Name:  "agentflow-turn",
	Title: "AgentFlow Turn",

	DescriptionLong:  "Executes the CLI Agent's assigned turn in an AgentFlow workspace. Use when the user says it is the CLI Agent's turn, asks to continue working on a feature, or mentions that the Web Reviewer has approved something. Trigger phrases include \"it's your turn\", \"the web reviewer approved\", \"continue the feature\", \"run your turn\", \"agentflow turn\", \"do the next step\".",
	DescriptionShort: "Executes the CLI Agent's turn in an AgentFlow workspace. Use this whenever the user says it is the CLI Agent's turn to act, or asks you to continue working on a feature.",

	Intro: "It is your turn to act. Read the current state of the exchange folder and do what is expected of you.",

	Idempotency: "Before executing any step, check what already exists in `discovery/`, `plans/`, `tasks/`, and `decisions/` that correspond to the current phase. If results already exist for the current phase, do not overwrite them — report what you find and ask the Human whether to continue or start a new task.\n\n" +
		"`plans/PLAN.md` is the one exception to \"do not overwrite\": there is always exactly one plan per feature. If it already exists and you're asked to plan again — because scope evolved, not because you're starting a new feature — edit it in place and add an entry to its `## Changelog` section describing what changed and why. Never create a second plan file (e.g. `plan-2.md`, `plan-v2.md`) next to it. If the new work is not a revision of the current plan's scope but something separable, stop and tell the Human this looks like a new feature rather than a plan revision.",

	Steps: []Step{
		{
			Title: "Pull latest changes from origin",
			Body:  "```bash\ngit pull --rebase\n```",
		},
		{
			Title: "Orient yourself",
			Body: "Read the latest file in `handoffs/` if it exists. Do this every time — it confirms context in ongoing sessions and re-orients you in new ones.\n\n" +
				"If you find a `handoffs/web-review-<date>-<slug>.md` file, that is the Web Reviewer telling you it could not write `STATUS.md`/`state.json` (the GitHub connector blocked the write). Do not treat it as authoritative on its own and do not sync `STATUS.md`/`state.json` from it automatically — tell the Human what it says and ask for explicit instruction before making any change based on it.",
		},
		{
			Title: "Read the exchange folder",
			Body: "Find: `.agentflow/features/*/`\n\n" +
				"Read in order:\n" +
				"1. `CONFIG.md` — feature scope, constraints, policies\n" +
				"2. `STATUS.md` — current phase, turn, next action\n" +
				"3. `state.json` — confirm `current_turn` is `cli`\n\n" +
				"If `current_turn` is not `cli`, stop and tell the Human:\n" +
				"> \"The current turn is not assigned to the CLI Agent. STATUS.md says: [current_turn value].\"\n\n" +
				"If there are files in `decisions/` that have not been resolved, stop and tell the Human:\n" +
				"> \"There is an unresolved decision request. Please resolve it before asking me to continue.\"",
		},
		{
			Title: "Read phase context",
			Body: "Read everything in the exchange folder that is relevant to the current phase:\n" +
				"- `specs/` — feature specifications\n" +
				"- `discovery/` — research and codebase analysis\n" +
				"- `plans/PLAN.md` — the current implementation plan (single file; check its `## Changelog` for revisions)\n" +
				"- `tasks/` — previous task results\n" +
				"- `reviews/` — feedback from the Web Reviewer",
		},
		{
			Title: "Do the work",
			Body: "Act according to what `STATUS.md` says is the `next_action`. Use your judgment — the context in the exchange folder is enough to know what is expected.\n\n" +
				"Core policies (from CONFIG.md — always apply):\n" +
				"- Do not implement anything before discovery and planning have been approved by the Web Reviewer.\n" +
				"- When you find ambiguity or risk, do not guess. Create a decision request instead (see next step).\n" +
				"- There is exactly one plan file: `plans/PLAN.md`. Create it once; on every later revision, edit it in place and add an entry under its `## Changelog` heading instead of creating a new file.\n" +
				"- Once Phase 3 (implementation) is approved — `STATUS.md` says phase `done`, next action mentions merge — that approval already covers the PR and the merge. Open the PR if repo policy requires one, merge it, then run `agentflow close`. Do not create a decision request or otherwise ask the Web Reviewer to re-approve the PR or the merged result; only escalate if something unexpected happens (conflicts, failing CI).\n" +
				"- The feature branch is done the moment the PR merges. Never push another commit to it afterward for any reason — not to record that it merged, not to update STATUS.md/state.json, not for a handoff note. That trailing commit is exactly what makes `agentflow close` refuse to delete the branch (it's no longer fully contained in the merge). If you need to note that the feature is closed, do it in a commit on the root branch, or simply let `agentflow close` remove the exchange folder.\n" +
				"- Write a result file at the end of every task.\n" +
				"- Update STATUS.md and state.json at every handoff.",
		},
		{
			Title: "If you find ambiguity or risk",
			Body: "Create a file at:\n```\ndecisions/decision-<YYYY-MM-DD>-<short-slug>.md\n```\n\n" +
				"With this content:\n```markdown\n" +
				"# Decision Request — <short title>\n\n" +
				"## Context\n<what you were doing when you found the issue>\n\n" +
				"## Question\n<the specific question that needs resolution>\n\n" +
				"## Options\n- Option A: ...\n- Option B: ...\n\n" +
				"## Recommendation\n<your recommendation if you have one>\n```\n\n" +
				"Then update STATUS.md and state.json with `current_turn: web` and stop. The Web Reviewer will read the decision file, discuss with the Human if needed, and return the turn to `cli`.",
		},
		{
			Title: "Write a result file",
			Body: "At the end of every task, write a result file at:\n```\ntasks/task-<YYYY-MM-DD>-<short-slug>.md\n```\n" +
				"Summarizing what you did, what files you created or modified, and any open questions.",
		},
		{
			Title: "Update STATUS.md and state.json",
			Body: "Update `STATUS.md`:\n```markdown\n" +
				"## Current phase\n<current or next phase>\n\n" +
				"## Current turn\nweb\n\n" +
				"## Status\n<short description of what was completed>\n\n" +
				"## Next action\n<what the Web Reviewer should do>\n\n" +
				"## Last update\n<timestamp>\n```\n\n" +
				"Update `state.json` with the same values (`current_phase`/`current_turn` must match `STATUS.md` literally; `status`/`next_action` are a machine-slug equivalent of the same fact, not a literal copy — see `docs/WEB-AGENT-ROLE.md` for the exact rule).",
		},
		{
			Title: "Write a handoff note",
			Body: "Write a handoff note at:\n```\nhandoffs/handoff-<YYYY-MM-DD>-{{AGENT_SLUG}}.md\n```\n" +
				"With this content:\n```markdown\n" +
				"# Handoff — <phase> — <date>\n\n" +
				"## Agent\n{{AGENT_NAME}}\n\n" +
				"## What was done\n<summary of this turn>\n\n" +
				"## Files created or modified\n<list>\n\n" +
				"## Next step for the CLI Agent\n<exact next action when the turn returns to cli>\n\n" +
				"## Open questions\n<if any, otherwise \"none\">\n```",
			ClaudeExtra: "Also copy the same content to `.claude/handoffs/LATEST.md` in the project root — Claude Code auto-injects this file as context at session start.",
		},
		{
			Title: "Commit and push",
			Body:  "```bash\ngit add -A\ngit commit -m \"cli: <short description of what was done>\"\ngit push\n```",
		},
	},

	Constraints: []string{
		"Do not start the next phase after finishing.",
		"Do not implement before planning is approved.",
		"One turn = one phase step. Stop when the handoff is done.",
	},

	SuccessCriteria: []string{
		"Result file exists in `tasks/`",
		"STATUS.md shows `current_turn: web`",
		"Changes committed and pushed",
	},

	NextStepMessage: "Done. The Web Reviewer can now review [what you did] on branch [branch name].",

	StopMessage: "Do not start the next phase or continue working. Your turn is complete.",
}
