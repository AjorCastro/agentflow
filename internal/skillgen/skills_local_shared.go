package skillgen

// Shared step groups for AgentFlow's local-mode skills. There are 5 skills
// (agentflow-local-init, agentflow-local-resume, agentflow-local-coder-init,
// agentflow-local-coder-resume, agentflow-local-pause), but only 2 distinct
// role loops: what the Controller does once it has a feature to work on,
// and what the Coder does once it has a task to work on. -init variants
// bootstrap into the same loop; -resume variants start directly in it. See
// .agentflow-local/discussions/001/{OUTCOME,PROTOTYPE-PLAN}.md for the
// design this implements.

// controllerLoadStep is how the Controller loads its bounded context,
// shared verbatim by -init (after bootstrap) and -resume.
var controllerLoadStep = Step{
	Title: "Load only the small artifacts",
	Body: "Run:\n```bash\nagentflow local context --role controller\n```\n" +
		"(the feature is whichever `agentflow local init` last recorded as current for this worktree — pass `--feature <id>` only if you need to target a different one). " +
		"This prints exactly `POLICY.md` + `checkpoint.md` + `task.md` + `result.md` (whichever exist) — nothing more. " +
		"Never read `history/` or `runtime/` under `.agentflow/local/<feature-id>/`: they are audit-only trails, never a source of truth, and reading them defeats the purpose of this mode (bounded context per round).",
}

// controllerDecideAndTaskSteps are the Controller's core round-trip: decide
// with the Human, write task.md. Shared verbatim by -resume; -init's
// bootstrap replaces these with its own "define the first task" step
// worded for a feature that has no history yet.
var controllerDecideAndTaskSteps = []Step{
	{
		Title: "Converse with the Human and decide the next step",
		Body:  "Use `checkpoint.md` (current state) and `result.md` (if a task just came back) to ground the conversation. Decide what happens next: a new task for the Coder, a decision the Human needs to make, or closing the feature.",
	},
	{
		Title: "Define the next task",
		Body: "Write `task.md` with the objective, acceptance criteria, scope constraints, and pointers to relevant files (not their full content). Keep it short enough that a Coder starting a brand-new session can act on it without asking you to re-explain anything already in `POLICY.md`/`checkpoint.md`.\n\n" +
			"If deciding what to ask for requires evidence from the repository (how something is currently implemented, whether a gap actually exists), delegate that investigation to a read-only sub-agent instead of reading the codebase yourself in this session — same reasoning as for the Coder (see `agentflow-local-coder-resume`): a specific question in, a short structured answer with file:line references out. Then instruct the Coder in `task.md` to do the same for whatever open-ended investigation their task still needs.\n\n" +
			"```bash\nagentflow local task <<'EOF'\n# Task\n\n## Objective\n...\n\n## Acceptance criteria\n- ...\n\n## Scope\n...\nEOF\n```",
	},
}

// controllerReviewAndLifecycleSteps are the tail shared by -init (after its
// bootstrap + first task) and -resume (after controllerDecideAndTaskSteps):
// review, checkpoint, deliberations, restart triggers.
var controllerReviewAndLifecycleSteps = []Step{
	{
		Title: "Review the Coder's result",
		Body:  "Read `result.md`. Check it against the acceptance criteria in the corresponding `task.md`. If something is missing or wrong, write a new `task.md` describing the fix — do not implement it yourself.",
	},
	{
		Title: "Checkpoint at a milestone",
		Body: "When a plan is approved, a task finishes, the Human redirects the goal, or the session has grown large, overwrite `checkpoint.md` — objective/status, last task+result summary, files touched, test/build status, open risks, non-obvious decisions, next step — before doing anything else:\n" +
			"```bash\nagentflow local checkpoint <<'EOF'\n...\nEOF\n```",
	},
	{
		Title: "Isolated human deliberations",
		Body: "For a decision that needs real back-and-forth with the Human (not a routine task assignment), start one:\n```bash\nagentflow local discuss start\n```\n" +
			"Deliberate. When you reach a conclusion, close it with a self-contained `OUTCOME.md` — conclusions only, never the transcript:\n" +
			"```bash\nagentflow local discuss close --id <id> <<'EOF'\n...\nEOF\n```\n" +
			"Closing a discussion is a session-restart trigger (see next step).",
	},
	{
		Title: "Restart your session at the right triggers",
		Body: "You must end this conversation and start a brand-new one — via `agentflow-local-resume`, reading only `POLICY.md`+`checkpoint.md`, plus `task.md`/`result.md` if a round is in flight — after any of:\n" +
			"- Closing a deliberation (`OUTCOME.md` just written).\n" +
			"- Approving a plan.\n" +
			"- The Human changes the feature's objective.\n" +
			"- This session's context has grown large enough that re-reading it is itself expensive.\n\n" +
			"If the Human needs you to stop for an unrelated reason (end of day, interruption) before any of these triggers fire, use `agentflow-local-pause` instead of just closing the session — it captures whatever partial progress exists so `agentflow-local-resume` doesn't start blind.\n\n" +
			"Before ending the session, make sure everything needed to resume is already in `checkpoint.md` — if you can't summarize it there, you're not at a valid restart point yet.",
		ClaudeExtra: "In Claude Code this means literally closing this chat and starting a new one with `/agentflow-local-resume` — there is no in-session \"soft reset\" that achieves the same bounded-context effect.",
	},
}

var localControllerConstraints = []string{
	"Never write or edit source code — that is the Coder's job.",
	"Never read history/ or runtime/ under .agentflow/local/<feature-id>/.",
	"Never let checkpoint.md become a duplicate of POLICY.md — POLICY.md is for what doesn't change; checkpoint.md is for what does.",
	"Never keep a deliberation's transcript once OUTCOME.md is written.",
}

var localControllerSuccessCriteria = []string{
	"task.md or a discussion OUTCOME.md reflects the Human's actual decision",
	"checkpoint.md is current enough that a fresh session could resume from it alone",
	"No source code was modified by this session",
}

// coderLoadStep is how the Coder loads its bounded context, shared
// verbatim by -init (after confirming the link) and -resume.
var coderLoadStep = Step{
	Title: "Load only the small artifacts",
	Body: "Run:\n```bash\nagentflow local context --role coder\n```\n" +
		"(the feature is whichever `agentflow local init` last recorded as current for this worktree — pass `--feature <id>` only if you need to target a different one). " +
		"This prints exactly `POLICY.md` + `checkpoint.md` + `task.md` — nothing more. " +
		"Do not read any other session's conversation, and never read `history/` or `runtime/` under `.agentflow/local/<feature-id>/`: they are audit-only, never a source of truth for what to do next.",
}

// coderWorkSteps are the Coder's core round-trip: understand, implement,
// report, checkpoint, stop. Shared verbatim by -init and -resume.
var coderWorkSteps = []Step{
	{
		Title: "Understand the task",
		Body:  "Read `task.md`'s objective, acceptance criteria, and scope. If something essential is missing or ambiguous and isn't resolved by `POLICY.md`/`checkpoint.md` either, stop and tell the Human what's missing instead of guessing — that gap should be added to `checkpoint.md` by the Controller before you continue.",
	},
	{
		Title: "Investigate, plan, implement, test",
		Body: "Do the work described in `task.md`, following whatever policies `POLICY.md` documents (validation gates, commit conventions, scope constraints). Use `git` normally for the actual code — branch, commits — local mode only changes the coordination channel, not how code is versioned.\n\n" +
			"Whenever the task requires open-ended exploration (reading unfamiliar code across several files, running an experiment just to learn a fact, searching for where something is defined) — delegate that to a sub-agent instead of doing it in this session directly. Ask it a specific question and have it report back a short, structured answer with file:line references or concrete evidence, not full file dumps. This keeps this session's own context small, which is the entire point of local mode's per-task session model — reading through half the codebase yourself defeats it just as surely as re-reading old conversation history would.",
	},
	{
		Title: "Write result.md",
		Body: "Summarize what you did: files touched (list, not full diff), test/build results, blockers, suggested next step. Keep it short — this is what the Controller (and a future fresh Coder session) will read instead of your conversation.\n\n" +
			"```bash\nagentflow local result <<'EOF'\n# Result\n\n## What was done\n...\n\n## Files touched\n...\n\n## Tests/build\n...\n\n## Blockers\n...\n\n## Suggested next step\n...\nEOF\n```",
	},
	{
		Title: "Checkpoint before ending the session",
		Body: "Finishing a task is always a session-end trigger for the Coder — overwrite `checkpoint.md` with the current goal/status, this task's outcome, files touched, test/build status, open risks, and next suggested step, so either the Controller or a brand-new Coder session can resume without you:\n" +
			"```bash\nagentflow local checkpoint <<'EOF'\n...\nEOF\n```",
	},
	{
		Title:       "Stop",
		Body:        "Once `result.md` and `checkpoint.md` are written, your session is done. Do not start another task or keep investigating ahead of what was asked. The next task is a new session — via `agentflow-local-coder-resume` — not a continuation of this one.",
		ClaudeExtra: "In Claude Code this means the Human closes this chat and opens a new one with `/agentflow-local-coder-resume` for the next task — carrying context forward in this same chat would defeat the bounded-context design of local mode.",
	},
}

var localCoderConstraints = []string{
	"Never continue past the one task assigned in task.md.",
	"Never read history/ or runtime/ under .agentflow/local/<feature-id>/.",
	"Never rely on memory of a previous task's session — only on what's written in POLICY.md/checkpoint.md/task.md.",
	"Never skip writing result.md, even if the task failed or was blocked.",
	"Never do open-ended exploration inline when a sub-agent could do it and report back a short answer instead.",
}

var localCoderSuccessCriteria = []string{
	"result.md exists and reflects what was actually done",
	"checkpoint.md is current enough for a fresh session to resume",
	"No unrelated work was done beyond the scope of task.md",
}
