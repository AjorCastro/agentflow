package skillgen

// LocalCoderResumeSkill is the single source of truth for the
// agentflow-local-coder-resume skill: the Coder's steady-state loop in
// AgentFlow's local mode, used for every task after the initial link
// confirmed by agentflow-local-coder-init. A brand-new session per task,
// per design. See .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md
// for the design this skill follows.
var LocalCoderResumeSkill = SkillDef{
	Name:  "agentflow-local-coder-resume",
	Title: "AgentFlow Local Coder Resume",

	DescriptionLong:  "Resumes as the Coder in AgentFlow's local mode, in a feature already linked by agentflow-local-coder-init: starts a clean session for exactly one task assigned by the Controller, investigates, plans, implements, tests, and reports back — then the session ends. Use for every Coder session after the first one. Trigger phrases include \"resume as Coder\", \"agentflow local coder resume\", \"execute the task\", \"it's your turn\" (in a local-mode context).",
	DescriptionShort: "Resumes as the Coder in an already-linked local-mode feature: executes exactly one task from task.md in a clean session, then reports back via result.md.",

	Intro: "You are the Coder in AgentFlow's local mode, already linked to this feature by a prior agentflow-local-coder-init session. You exist for exactly one task in this session. Load only the small artifacts below, do the work, write result.md, and stop — the next task will be a brand-new session, not a continuation of this one.",

	Idempotency: "Check `checkpoint.md`'s \"Last task and result\" section before starting: if it already describes the task in the current `task.md` as done, stop and tell the Human instead of redoing it.",

	Steps: append([]Step{coderLoadStep}, coderWorkSteps...),

	Constraints:     localCoderConstraints,
	SuccessCriteria: localCoderSuccessCriteria,

	NextStepMessage: "Task complete. Signal the Controller session that result.md is ready for review.",

	StopMessage: "Do not start another task. This session is done — the next task begins a new one, via agentflow-local-coder-resume.",
}
