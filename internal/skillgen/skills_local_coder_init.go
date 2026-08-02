package skillgen

// LocalCoderInitSkill is the single source of truth for the
// agentflow-local-coder-init skill: the Coder's very first session for a
// feature, confirming the link the Controller just set up
// (agentflow-local-init) before doing the first task. Every task after this
// one uses agentflow-local-coder-resume instead. See
// .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md for the
// design this skill follows.
var LocalCoderInitSkill = SkillDef{
	Name:  "agentflow-local-coder-init",
	Title: "AgentFlow Local Coder Init",

	DescriptionLong:  "Links a new Coder session to a feature the Controller just bootstrapped with agentflow-local-init, confirms the link, then executes the first task. Use exactly once per feature, right after the Human opens this session per the Controller's instructions. Trigger phrases include \"link as Coder\", \"agentflow local coder init\", \"start the Coder for this feature\".",
	DescriptionShort: "Links a new Coder session to a feature the Controller just bootstrapped, then executes the first task.",

	Intro: "You are the Coder in AgentFlow's local mode, starting for the very first time on this feature. Before doing any work, confirm you're linked to the right feature — the Controller should have already run agentflow-local-init in a separate session.",

	Idempotency: "If `checkpoint.md` already describes a completed task, this feature isn't actually new to the Coder — use agentflow-local-coder-resume instead.",

	Steps: append([]Step{
		{
			Title: "Confirm you're linked to the right feature",
			Body: "Run (from inside the feature's worktree):\n```bash\nagentflow local context --role coder\n```\n" +
				"The feature is auto-detected from this worktree's current feature, set by the Controller's `agentflow local init`. If this errors with \"no --feature given and no current feature set\", the Controller hasn't bootstrapped yet in this worktree — stop and tell the Human instead of guessing a feature ID.\n\n" +
				"If it succeeds, confirm `POLICY.md` looks real (not placeholder text) and `task.md` exists — that's the first task waiting for you.",
		},
	}, coderWorkSteps...),

	Constraints:     localCoderConstraints,
	SuccessCriteria: append([]string{
		"Confirmed the link to the correct feature before doing any work",
	}, localCoderSuccessCriteria...),

	NextStepMessage: "Task complete. Signal the Controller session that result.md is ready for review.",

	StopMessage: "Do not start another task. This session is done — the next task begins a new one, via agentflow-local-coder-resume.",
}
