package skillgen

// LocalInitSkill is the single source of truth for the agentflow-local-init
// skill: bootstraps AgentFlow's local mode for a new feature in two stages —
// Stage A (pre-worktree: problem analysis, impact assessment, a validated
// PLAN.md, Human go/no-go — all from the root repo checkout, no worktree
// yet) and Stage B (worktree, coordination folder promoted into it,
// POLICY.md, first task) — then falls straight into the same steady-state
// loop as agentflow-local-resume, since nothing about having just
// bootstrapped requires a session restart. See
// .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md for the
// original design this skill follows, and the Human's 2026-08-03 feedback
// (no discovery/plan equivalent existed before this) for why Stage A exists.
var LocalInitSkill = SkillDef{
	Name:  "agentflow-local-init",
	Title: "AgentFlow Local Init (Controller bootstrap)",

	DescriptionLong:  "Bootstraps AgentFlow's local mode for a brand-new feature: analyzes the problem/need with the Human, assesses impact, drafts and validates a PLAN.md, then — once approved — creates the worktree, promotes the coordination folder into it, drafts POLICY.md, and assigns the first task before continuing as the Controller. Use once per feature, right after the Human decides in conversation that local mode work should start. Trigger phrases include \"start local mode\", \"agentflow local init\", \"bootstrap a local feature\", \"set up Controller/Coder for this feature\".",
	DescriptionShort: "Bootstraps a brand-new local-mode feature: analysis, impact, a validated PLAN.md, Human approval, then worktree + POLICY.md + first task.",

	Intro: "You are bootstrapping AgentFlow's local mode for a new feature, as the Controller. This runs once per feature — every session after this one uses agentflow-local-resume instead. It has two stages: Stage A happens in the root repo checkout, before any worktree exists, and ends with the Human approving a PLAN.md. Stage B only starts after that approval.",

	Idempotency: "If `.agentflow/local/<feature-id>/` already exists inside a worktree for this feature, do not re-run this skill — use agentflow-local-resume instead. If a *pending* `.agentflow/local/<feature-id>/` already exists in the root repo (Stage A was started in an earlier session), resume Stage A from wherever `PLAN.md`/`POLICY.md` left off instead of starting over.",

	Steps: append([]Step{
		{
			Title: "Confirm scope with the Human",
			Body: "By the time this skill runs, the Human should already have decided, in conversation, that a feature will start and roughly what it's for. If that conversation hasn't happened yet, have it first — this skill is not a substitute for deciding scope.\n\n" +
				"**Fast track exception:** if the Human explicitly asks for it, the fix is small and well understood (a handful of files, no schema/API/contract changes, nothing that needs analysis to even understand), and getting it wrong is low-risk and easily reversible, skip Stage A entirely and go straight to Stage B with a single task — no `PLAN.md`. You never decide to fast-track on your own; the Human must ask for it. If it turns out mid-work to be bigger than expected, stop and ask the Human to switch to the full flow instead of improvising.",
		},
		{
			Title: "Bootstrap the pending coordination folder",
			Body: "Run this from the **root repo checkout** — there is no worktree yet:\n```bash\nagentflow local init --repo <root-repo-path> --feature <name> --root <root-branch>\n```\n" +
				"This creates `.agentflow/local/<name>/` (POLICY.md skeleton, empty checkpoint.md, required subfolders) right there in the root checkout, and records `<name>` as the root repo's current feature. It moves into the worktree at the end of Stage B — nothing about this location is permanent.",
		},
		{
			Title: "Draft POLICY.md's stable questions",
			Body: "Open `.agentflow/local/<name>/POLICY.md` and fill in the placeholders directly — it's hand-authored, not piped through a CLI command like checkpoint.md/task.md/result.md/PLAN.md are. At minimum:\n" +
				"- **Validation gate**: check this repo's `AGENTS.md`/`CONTRIBUTING.md` for mandatory test/build commands or a CI/promotion gate; ask the Human if none is documented.\n" +
				"- **Commit conventions**: this feature's conventions beyond the repo's defaults, if any (the generated skeleton already has a fixed \"Commit discipline\" section — this is about anything additional).\n\n" +
				"If evidence from the repo is needed to answer either question, delegate that lookup to a read-only sub-agent instead of exploring the codebase yourself in this session. None of this depends on the worktree existing, so it happens now.",
		},
		{
			Title: "Analyze the problem or need",
			Body:  "Talk it through with the Human: what problem or need is this solving, and why now? Capture this as PLAN.md's first section — a short problem statement, not a transcript.",
		},
		{
			Title: "Assess impact",
			Body: "Delegate to a read-only sub-agent: which files/modules are affected, a rough size-of-effort estimate, and any risk areas (schema/API/contract changes, code with poor test coverage, anything touching shared infrastructure). Give it a specific question and have it report back a short, structured answer with `file:line` references — never full file dumps or an open-ended \"explore the codebase\" instruction. Do not do this exploration yourself in this session.",
		},
		{
			Title: "Draft PLAN.md",
			Body: "Write PLAN.md with three sections: Problem/Need (from the analysis step), Impact assessment (from the sub-agent), and a Work plan — a sequenced checklist, one line per step, GFM task-list syntax so progress can be tracked automatically:\n" +
				"```bash\nagentflow local plan <<'EOF'\n# Plan\n\n## Problem/Need\n...\n\n## Impact assessment\n- Affected files/modules: ...\n- Size of effort: ...\n- Risks: ...\n\n## Work plan\n- [ ] <first step, short objective>\n- [ ] <second step>\n...\nEOF\n```\n" +
				"Each checklist item should be independently completable and independently committable — this is what makes Stage B's task-by-task iteration and partial commits possible.",
		},
		{
			Title: "Validate the plan from an implementer's perspective",
			Body: "Delegate to a sub-agent explicitly briefed to review PLAN.md the way the Coder would when handed each step cold: Is the sequencing right? Is any step's acceptance criteria vague enough that an implementer would have to guess? Does the impact assessment actually match what the plan proposes to touch? Is anything missing?\n\n" +
				"If the review surfaces real gaps, revise and re-run `agentflow local plan` (the previous version is archived to `history/` automatically) before moving on — do not present an unreviewed plan to the Human.",
		},
		{
			Title: "Get the Human's go/no-go",
			Body: "Present PLAN.md to the Human.\n" +
				"- **Approved** → move to Stage B below.\n" +
				"- **Changes requested** → revise PLAN.md (repeat the draft/validate steps above) and present again.\n" +
				"- **No-go** → stop here. Nothing to clean up — the pending folder stays harmlessly in the root repo's gitignored `.agentflow/local/` until the Human revisits it or it's removed later.",
		},
		{
			Title: "Create the worktree",
			Body: "If the feature's worktree/branch don't exist yet, create them:\n```bash\ngit worktree add .worktrees/<name> -b feature/<name> <root-branch>\ncd .worktrees/<name>\n```\n" +
				"If the Human already created them, just `cd` into the worktree and verify with `git worktree list`. All following commands run from inside that worktree.",
		},
		{
			Title: "Promote the pending coordination folder",
			Body: "Move Stage A's folder (POLICY.md, checkpoint.md, PLAN.md, history/) from the root repo into this worktree:\n```bash\nagentflow local promote --feature <name> --from <root-repo-path>\n```\n" +
				"(run from inside the worktree — `--repo` defaults to `.`). This also records `<name>` as this worktree's current feature, so every later `agentflow local` command here can omit `--feature`.",
		},
		{
			Title: "Fill in POLICY.md's branch/worktree/root fields",
			Body:  "These were unknown during Stage A (no worktree existed yet). Edit `.agentflow/local/<name>/POLICY.md` by hand now that `promote`'s output has reminded you they're known: `root_branch`, `branch`, `worktree`.",
		},
		{
			Title: "Seed the first task from PLAN.md",
			Body: "Write task.md from PLAN.md's **first** unchecked checklist item — objective, acceptance criteria, scope constraints, and pointers to the specific files the impact assessment surfaced (not their full content), so the Coder isn't repeating that exploration from scratch.\n" +
				"```bash\nagentflow local task <<'EOF'\n# Task\n\n## Objective\n...\n\n## Acceptance criteria\n- ...\n\n## Scope\n...\nEOF\n```",
		},
		{
			Title: "Tell the Human to link the Coder",
			Body:  "The Human should now open a second, independent session in the same worktree and invoke `agentflow-local-coder-init` there. Give them the exact instruction to do so before continuing — this is the point at which Controller and Coder become linked and ready to work.",
		},
	}, controllerReviewAndLifecycleSteps...),

	Constraints: localControllerConstraints,
	SuccessCriteria: append([]string{
		"PLAN.md (when this wasn't fast-tracked) reflects a Human-approved problem statement, impact assessment, and sequenced checklist — not placeholder text",
		"POLICY.md reflects real repo conventions, not placeholder text",
		"task.md exists and describes real, actionable first work drawn from PLAN.md's first checklist item",
	}, localControllerSuccessCriteria...),

	NextStepMessage: "Bootstrap complete and first task assigned. Tell the Human to open a second session and run agentflow-local-coder-init there.",

	StopMessage: "Do not implement the task yourself. Wait for the Human to signal that the Coder has written result.md.",
}
