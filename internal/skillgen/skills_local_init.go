package skillgen

// LocalInitSkill is the single source of truth for the agentflow-local-init
// skill: bootstraps AgentFlow's local mode for a new feature — worktree,
// coordination folder, POLICY.md, first task — then falls straight into the
// same steady-state loop as agentflow-local-resume, since nothing about
// having just bootstrapped requires a session restart. See
// .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md for the
// design this skill follows.
var LocalInitSkill = SkillDef{
	Name:  "agentflow-local-init",
	Title: "AgentFlow Local Init (Controller bootstrap)",

	DescriptionLong:  "Bootstraps AgentFlow's local mode for a brand-new feature: creates the worktree if needed, runs `agentflow local init`, drafts POLICY.md with the Human, and assigns the first task — then continues as the Controller. Use once per feature, right after the Human decides in conversation that local mode work should start. Trigger phrases include \"start local mode\", \"agentflow local init\", \"bootstrap a local feature\", \"set up Controller/Coder for this feature\".",
	DescriptionShort: "Bootstraps a brand-new local-mode feature (worktree, coordination folder, POLICY.md, first task), then continues as Controller.",

	Intro: "You are bootstrapping AgentFlow's local mode for a new feature, as the Controller. This only runs once per feature — every session after this one uses agentflow-local-resume instead.",

	Idempotency: "If `.agentflow/local/<feature-id>/` already exists for this feature, do not re-run this skill — use agentflow-local-resume instead. This skill is only for a feature that has never been bootstrapped.",

	Steps: append([]Step{
		{
			Title: "Confirm scope with the Human",
			Body:  "By the time this skill runs, the Human should already have decided, in conversation, that a feature will start and roughly what it's for. If that conversation hasn't happened yet, have it first — this skill is not a substitute for deciding scope.",
		},
		{
			Title: "Create or verify the worktree",
			Body: "If the feature's worktree/branch don't exist yet, create them:\n```bash\ngit worktree add .worktrees/<name> -b feature/<name> <root-branch>\ncd .worktrees/<name>\n```\n" +
				"If the Human already created them, just `cd` into the worktree and verify with `git worktree list`. All following commands in this skill run from inside that worktree.",
		},
		{
			Title: "Run agentflow local init",
			Body: "```bash\nagentflow local init --feature <name> --branch feature/<name> --worktree .worktrees/<name> --root <root-branch>\n```\n" +
				"This creates `.agentflow/local/<name>/` (POLICY.md skeleton, empty checkpoint.md, required subfolders) and records `<name>` as this worktree's current feature — every `agentflow local` command run from this worktree afterwards can omit `--feature`.",
		},
		{
			Title: "Draft POLICY.md with the Human",
			Body: "Open `.agentflow/local/<name>/POLICY.md` and fill in the placeholders directly — it's hand-authored, not piped through a CLI command like checkpoint.md/task.md/result.md are. At minimum:\n" +
				"- **Validation gate**: check this repo's `AGENTS.md`/`CONTRIBUTING.md` for mandatory test/build commands or a CI/promotion gate; ask the Human if none is documented.\n" +
				"- **Commit conventions**: this feature's conventions beyond the repo's defaults, if any.\n\n" +
				"If evidence from the repo is needed to answer either question, delegate that lookup to a read-only sub-agent instead of exploring the codebase yourself in this session.",
		},
		{
			Title: "Define the first task",
			Body: "Same as any other task definition: objective, acceptance criteria, scope constraints, pointers to relevant files.\n" +
				"```bash\nagentflow local task <<'EOF'\n# Task\n\n## Objective\n...\n\n## Acceptance criteria\n- ...\n\n## Scope\n...\nEOF\n```",
		},
		{
			Title: "Tell the Human to link the Coder",
			Body:  "The Human should now open a second, independent session in the same worktree and invoke `agentflow-local-coder-init` there. Give them the exact instruction to do so before continuing — this is the point at which Controller and Coder become linked and ready to work.",
		},
	}, controllerReviewAndLifecycleSteps...),

	Constraints:     localControllerConstraints,
	SuccessCriteria: append([]string{
		"POLICY.md reflects real repo conventions, not placeholder text",
		"task.md exists and describes real, actionable first work",
	}, localControllerSuccessCriteria...),

	NextStepMessage: "Bootstrap complete and first task assigned. Tell the Human to open a second session and run agentflow-local-coder-init there.",

	StopMessage: "Do not implement the task yourself. Wait for the Human to signal that the Coder has written result.md.",
}
