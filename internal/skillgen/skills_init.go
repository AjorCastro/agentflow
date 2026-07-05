package skillgen

// InitSkill is the single source of truth for the agentflow-init skill
// across all 3 agent flavors.
var InitSkill = SkillDef{
	Name:  "agentflow-init",
	Title: "AgentFlow Init",

	DescriptionLong: "Reads agentflow-init-<branch-slug>.md created by the Web Reviewer and initializes the AgentFlow workspace on a new feature branch with worktree and exchange folder. Use when the user says the Web Reviewer has created agentflow-init-<branch-slug>.md, or asks to initialize a feature, create a worktree, or run agentflow init. Trigger phrases include \"the web reviewer created agentflow-init-<branch-slug>.md\", \"initialize the workspace\", \"run agentflow init\", \"create the feature branch\".",
	DescriptionShort: "Reads agentflow-init-<branch-slug>.md created by the Web Reviewer and initializes the AgentFlow workspace. Use this when the user says the Web Reviewer has created agentflow-init-<branch-slug>.md and asks you to initialize the workspace.",

	Intro: "Read the `agentflow-init-<branch-slug>.md` file created by the Web Reviewer and initialize the AgentFlow workspace.",

	Idempotency: "Before executing any step, check if it was already done. Skip completed steps silently and continue from where things left off. Never duplicate work or overwrite existing results. If everything is already done, say so and stop.",

	Steps: []Step{
		{
			Title: "Pull latest changes from origin",
			Body:  "The Web Reviewer created `agentflow-init-<branch-slug>.md` directly on GitHub. Pull to get it locally:\n\n```bash\ngit pull --rebase\n```",
		},
		{
			Title: "Find and read the agentflow-init file",
			Body: "Look for `agentflow-init-*.md` files in the repository root (current directory, on the root branch).\n\n" +
				"- If no file is found after pulling, stop and tell the Human:\n" +
				"  > \"No agentflow-init-*.md file found after git pull. Ask the Web Reviewer to confirm they committed it to the develop branch.\"\n" +
				"- If exactly one file is found, use it.\n" +
				"- If multiple files are found, list them and ask the Human which feature to initialize.\n\n" +
				"Read the file and extract these parameters:\n" +
				"- `title`\n- `branch`\n- `worktree`\n- `root` (default: `develop` if not specified)",
		},
		{
			Title: "Verify you are on the root branch",
			Body: "Check the current branch:\n```bash\ngit rev-parse --abbrev-ref HEAD\n```\n\n" +
				"If you are not on the root branch (usually `develop`), stop and tell the Human which branch you are on.",
		},
		{
			Title: "Remove the agentflow-init file from the root branch BEFORE init",
			Body: "This is critical: the init file must be removed from `develop` before the worktree is created, otherwise the file will be inherited by the new feature branch.\n\n" +
				"Let `<init-file>` be the filename found in the previous step.\n\n" +
				"Check if it still exists:\n" +
				"- If it exists → remove and commit on the root branch:\n" +
				"  ```bash\n  git rm <init-file>\n  git commit -m \"chore: consume <init-file>\"\n  git push\n  ```\n" +
				"- If it was already removed → skip.",
		},
		{
			Title: "Run agentflow init",
			Body: "Check if the workspace already exists by running `agentflow status`.\n" +
				"- If the exchange folder for this branch already exists → skip to the verification step.\n" +
				"- If it does not exist → run:\n\n" +
				"```bash\nagentflow init \\\n  --root <root> \\\n  --branch <branch> \\\n  --worktree <worktree> \\\n  --title \"<title>\"\n```\n\n" +
				"The `--push` flag is enabled by default. If there is no remote configured yet, `agentflow init` will fail with a clear message — tell the Human to add a remote first:\n" +
				"```bash\ngit remote add origin git@github.com:<user>/<repo>.git\n```",
		},
		{
			Title: "Verify the result",
			Body: "Run:\n```bash\nagentflow status\n```\n\n" +
				"Confirm that:\n- The exchange folder was created\n- `phase` is `intake`\n- `turn` is `web`\n- `next_action` is `begin_discovery`",
		},
	},

	Constraints: []string{
		"Do not begin discovery or any other task after init.",
		"Do not change branches in the main repo.",
		"The next turn belongs to the Web Reviewer.",
	},

	SuccessCriteria: []string{
		"`agentflow status` shows exchange folder with `phase: intake`, `turn: web`, `next_action: begin_discovery`",
		"The `agentflow-init-*.md` file no longer exists in the repo root",
		"Branch and worktree are pushed to GitHub",
	},

	SummaryTemplate: "Branch   : <branch>\n" +
		"Worktree : <worktree>\n" +
		"Exchange : <exchange folder path>\n" +
		"Status   : initialized\n\n" +
		"Next step: Tell the Web Reviewer the workspace is ready on branch <branch>.\n" +
		"           They can now read the exchange folder on GitHub and begin Phase 1 — Discovery.",

	NextStepMessage: "Workspace ready on branch <branch>. Tell the Web Reviewer they can read the exchange folder on GitHub and begin Phase 1 — Discovery.",

	StopMessage: "Do not begin discovery or any other task. The next turn belongs to the Web Reviewer.",
}
