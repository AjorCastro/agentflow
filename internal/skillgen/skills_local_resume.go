package skillgen

// LocalResumeSkill is the single source of truth for the
// agentflow-local-resume skill: the Controller's steady-state loop in
// AgentFlow's local mode, used for every session after the feature has
// already been bootstrapped by agentflow-local-init. See
// .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md for the
// design this skill follows.
var LocalResumeSkill = SkillDef{
	Name:  "agentflow-local-resume",
	Title: "AgentFlow Local Resume (Controller)",

	DescriptionLong:  "Resumes as the Controller in AgentFlow's local mode, in a feature already bootstrapped by agentflow-local-init: talks to the Human, defines the Coder's next task, reviews its results, and manages isolated human deliberations. Never implements code. Use for every Controller session after the first one — including forced session restarts mid-feature. Trigger phrases include \"resume as Controller\", \"agentflow local resume\", \"assign the next task\", \"review the Coder's result\", \"let's deliberate\".",
	DescriptionShort: "Resumes as the Controller in an already-bootstrapped local-mode feature: defines tasks for the Coder, reviews results, never implements code.",

	Intro: "You are the Controller in AgentFlow's local mode, resuming work on a feature that agentflow-local-init already bootstrapped. You talk to the Human, define tasks, take and record decisions, and review the Coder's results. You never implement code yourself — that is the Coder's job, in a separate session.",

	Idempotency: "Before doing anything else, check whether `task.md` or `result.md` already exist for the current round via `agentflow local context --role controller`. If a `result.md` is already waiting for review, review it before defining a new task — do not overwrite `task.md` mid-round.",

	Steps: append(
		append([]Step{controllerLoadStep}, controllerDecideAndTaskSteps...),
		controllerReviewAndLifecycleSteps...,
	),

	Constraints:     localControllerConstraints,
	SuccessCriteria: localControllerSuccessCriteria,

	NextStepMessage: "Task assigned. Signal the Coder session that a task is ready — a new one via agentflow-local-coder-init, or the next round via agentflow-local-coder-resume if it's already linked.",

	StopMessage: "Do not implement the task yourself. Wait for the Human to signal that the Coder has written result.md.",
}
