package skillgen

// SetupSkill is the single source of truth for the agentflow-setup skill
// across all 3 agent flavors.
//
// Step 4 ("Create docs/WEB-AGENT-ROLE.md") used to hand-embed the Web
// Reviewer role document's markdown directly in each flavor's skill file —
// 3 separate copies that drifted out of sync with each other and with the
// canonical content the binary actually generates (protocol.WebAgentRoleMD).
// It now just tells the agent to run `agentflow docs sync`, which writes
// that file from the binary's current canonical content. There is nothing
// left to keep in sync by hand.
var SetupSkill = SkillDef{
	Name:  "agentflow-setup",
	Title: "AgentFlow Setup",

	DescriptionLong: "Initializes a new software project for AgentFlow collaboration. Use when the user asks to set up a new project from scratch, create a git repository, publish to GitHub, or prepare a repo so a Web Reviewer can start working with AgentFlow. Trigger phrases include \"set up a new project\", \"initialize agentflow\", \"create the repo\", \"prepare for agentflow\".",
	DescriptionShort: "Initializes a new software project with git, publishes it to GitHub, and creates the AgentFlow Web Reviewer instructions file. Use this when the user asks to set up a new project or start from scratch with AgentFlow.",

	Intro: "Initialize a new project repository and publish it so the Web Reviewer can start working.",

	Idempotency: "Before executing any step, check if it was already done. Skip completed steps silently and continue from where things left off. Never duplicate work or overwrite existing results. If everything is already done, say so and stop.",

	HasSeparatePrereqSection: true,

	Steps: []Step{
		{
			Title: "Verify prerequisites",
			Body: "Check that the following tools are available:\n" +
				"- `git` — for version control\n" +
				"- `gh` — GitHub CLI, authenticated (`gh auth status`)\n" +
				"- `agentflow` — AgentFlow CLI\n\n" +
				"If any is missing, stop and tell the Human what needs to be installed before continuing.",
		},
		{
			Title: "Get project name",
			Body: "If the current directory name is not a suitable project name, ask the Human to confirm before proceeding.\n\n" +
				"Use the current directory name as the repository name unless the Human specifies otherwise.",
		},
		{
			Title: "Initialize git",
			Body: "Check if a git repository already exists (`git rev-parse --git-dir`).\n" +
				"- If it does not exist → run `git init -b develop`\n" +
				"- If it already exists → skip, verify current branch is `develop`. If not, tell the Human and stop.",
		},
		{
			Title: "Create docs/WEB-AGENT-ROLE.md",
			Body: "Run:\n```bash\nagentflow docs sync\n```\n" +
				"This writes (or refreshes) `docs/WEB-AGENT-ROLE.md` on the `develop` branch directly from the canonical content built into the `agentflow` binary, and commits it if it changed.\n\n" +
				"There is nothing to write by hand here — never hand-copy this file's content into a skill file or a chat message. A hand-copied version will drift out of date the next time the binary changes; `agentflow docs sync` never can, because it always reflects whatever binary is currently installed.",
		},
		{
			Title: "Create .gitignore",
			Body: "Check if `.worktrees/` is already ignored.\n" +
				"- If not → create or append `.worktrees/` to `.gitignore`.\n" +
				"- If yes → skip.",
		},
		{
			Title: "Commit",
			Body: "Check if there are staged or unstaged changes (`git status --short`).\n" +
				"- If there are changes → stage and commit:\n" +
				"  ```bash\n  git add -A\n  git commit -m \"chore: initial project setup\"\n  ```\n" +
				"- If the tree is already clean → skip.",
		},
		{
			Title: "Create the GitHub repository",
			Body: "Check if a remote named `origin` already exists (`git remote`).\n" +
				"- If it does not exist → create the repo and push:\n" +
				"  ```bash\n  gh repo create <project-name> --public --source . --remote origin --push\n  ```\n" +
				"  If the Human wants a private repository, use `--private`. When in doubt, ask.\n" +
				"- If `origin` already exists → check if the branch is up to date. If not, push. If already pushed → skip.",
		},
	},

	Constraints: []string{
		"Never overwrite existing files.",
		"Never change the current branch without explicit user instruction.",
		"Ask before creating a public repository when in doubt.",
	},

	SuccessCriteria: []string{
		"`git remote -v` shows `origin` pointing to GitHub",
		"`docs/WEB-AGENT-ROLE.md` exists and is committed",
		"`develop` branch is pushed to GitHub",
	},

	SummaryTemplate: "Project   : <project-name>\n" +
		"Branch    : develop\n" +
		"Remote    : <github-url>\n" +
		"Next step : Share the GitHub URL with the Web Reviewer and ask them to read docs/WEB-AGENT-ROLE.md",

	NextStepMessage: "Project is ready at <github-url>. Share this URL with the Web Reviewer and ask them to read docs/WEB-AGENT-ROLE.md to start.",

	StopMessage: "Do not continue with any other task. Wait for the Human to coordinate with the Web Reviewer.",
}
