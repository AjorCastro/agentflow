package skillgen

// CloseSkill is the single source of truth for the agentflow-close skill
// across all 3 agent flavors.
var CloseSkill = SkillDef{
	Name:  "agentflow-close",
	Title: "AgentFlow Close",

	DescriptionLong: "Closes an AgentFlow feature workspace after the branch has been merged to the root branch. Use when the user says a feature is done, merged, or completed and asks to clean up, remove the worktree, or delete the branch. Trigger phrases include \"the feature is merged\", \"close the feature\", \"clean up the workspace\", \"delete the branch\", \"agentflow close\".",
	DescriptionShort: "Closes an AgentFlow feature workspace after the branch has been merged. Use this when the user says a feature has been merged and asks to clean up the workspace.",

	Intro: "The feature has been merged. Clean up the workspace.",

	Idempotency: "Before executing any step, check if it was already done. If the worktree is already removed or the branch already deleted, skip those steps silently and continue. If everything is already clean, say so and stop.",

	Steps: []Step{
		{
			Title: "Read the current state",
			Body:  "Run `agentflow status` to find `feature_branch`, `worktree`, and `root_branch`.\n\nIf no AgentFlow workspace is found, stop and tell the Human.",
		},
		{
			Title: "Confirm the feature is merged",
			Body:  "Ask the Human to confirm that the branch has been merged to the root branch before continuing. Do not proceed without this confirmation.",
		},
		{
			Title: "Run agentflow close",
			Body: "Run this from the repository checked out on `root_branch` (not from the worktree being removed):\n\n" +
				"```bash\nagentflow close --branch <feature_branch> --root <root_branch>\n```\n\n" +
				"If the Human also wants to delete the remote branch:\n" +
				"```bash\nagentflow close --branch <feature_branch> --root <root_branch> --delete-remote\n```\n\n" +
				"If the branch is not detected as merged but the Human is certain it was merged, use `--force` — but never use it unless the Human explicitly asks for it:\n" +
				"```bash\nagentflow close --branch <feature_branch> --root <root_branch> --force\n```\n\n" +
				"`agentflow close` also removes `.agentflow/features/<feature-id>/` from `root_branch` if present — the merge carries those files into root, and leaving them there would clutter it permanently.",
		},
	},

	Constraints: []string{
		"Never delete a branch without explicit merge confirmation from the Human.",
		"Never use `--force` unless the Human explicitly requests it.",
	},

	SuccessCriteria: []string{
		"`git worktree list` no longer shows the feature worktree",
		"`git branch` no longer shows the feature branch",
		"`agentflow status` shows no workspace for that feature",
	},

	SummaryTemplate: "Feature  : <feature_branch>\n" +
		"Worktree : removed\n" +
		"Branch   : deleted\n" +
		"Exchange : removed from <root_branch>\n\n" +
		"Next step: Run /agentflow-setup or agentflow init to start a new feature.",

	NextStepMessage: "Feature <branch> closed. Run /agentflow-setup or agentflow init to start a new feature.",

	StopMessage: "Do not start any new work.",
}
