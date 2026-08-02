package skillgen

// LocalPauseSkill is the single source of truth for the agentflow-local-pause
// skill: a safe way to stop a Controller or Coder session off a natural
// milestone (end of day, an unrelated interruption) without losing the
// in-progress work a normal checkpoint wouldn't yet capture. Usable from
// either role's active session — the action is the same either way. See
// .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md for the
// design this skill follows.
var LocalPauseSkill = SkillDef{
	Name:  "agentflow-local-pause",
	Title: "AgentFlow Local Pause",

	DescriptionLong:  "Safely stops the current Controller or Coder session in AgentFlow's local mode before a natural milestone, by writing a partial checkpoint.md that captures in-progress work, so the next agentflow-local-resume or agentflow-local-coder-resume doesn't start blind. Use when the Human needs to interrupt a session for a reason unrelated to task/plan completion — end of day, an external interruption. Trigger phrases include \"pause here\", \"agentflow local pause\", \"I need to stop for now\", \"let's continue later\".",
	DescriptionShort: "Safely stops the current session (Controller or Coder) off a natural milestone by writing a partial checkpoint first.",

	Intro: "The Human needs to stop this session now, but you're not at one of the normal restart/end triggers (task done, plan approved, deliberation closed). Do not just end the conversation — write down what's true right now so the next session isn't starting blind.",

	Idempotency: "If you're already at a normal milestone (task just finished, plan just approved), use the ordinary checkpoint step from your role's skill instead — this skill is only for stopping mid-work.",

	Steps: []Step{
		{
			Title: "Do not reload context",
			Body:  "You're pausing an active session, not starting one — everything you need is already in this conversation. Do not re-run `agentflow local context`.",
		},
		{
			Title: "Write a partial checkpoint.md",
			Body: "Overwrite `checkpoint.md`, explicitly marked as a mid-work pause (not a completed milestone), covering: overall goal/status, exactly what's done and what's still in progress in the current task/round, files touched so far, test/build status if known, open blockers, and the precise next step — specific enough that a resumed session doesn't have to guess or redo work you already did.\n" +
				"```bash\nagentflow local checkpoint <<'EOF'\n# Checkpoint\n\n## Goal and overall status\n... (note: PAUSED mid-work, not a natural milestone) ...\n\n## Last task and result\n... (in progress: what's done, what's left) ...\n\n## Files touched\n...\n\n## Test/build status\n...\n\n## Open blockers or risks\n...\n\n## Next suggested step\n...\nEOF\n```",
		},
		{
			Title: "If you are the Coder and the task is genuinely incomplete, say so",
			Body:  "Do not write `result.md` for an unfinished task — result.md means the task is done. The checkpoint alone is what carries a paused, in-progress task forward.",
		},
		{
			Title: "Tell the Human it's safe to close this session",
			Body:  "Confirm the checkpoint was written, then stop. The Human resumes later via agentflow-local-resume (Controller) or agentflow-local-coder-resume (Coder) — never a continuation of this session.",
		},
	},

	Constraints: []string{
		"Never end a mid-work session without writing a partial checkpoint first.",
		"Never write result.md for an incomplete task just to have something to show.",
		"Never treat a pause as a normal milestone checkpoint — say explicitly in checkpoint.md that this was a pause, not a completion.",
	},

	SuccessCriteria: []string{
		"checkpoint.md accurately reflects in-progress, incomplete work — not a finished round",
		"A fresh resume session could pick up exactly where this one stopped, without re-asking the Human what was happening",
	},

	NextStepMessage: "Paused. Resume later with agentflow-local-resume or agentflow-local-coder-resume, depending on which role this was.",

	StopMessage: "Session ends here. Do not continue working after writing the pause checkpoint.",
}
