# AgentFlow Init

Read the `agentflow-init.md` file created by the Web Reviewer and initialize the AgentFlow workspace.

## Steps

### 1. Find and read agentflow-init.md

Look for `agentflow-init.md` in the repository root (current directory).

If the file does not exist, stop and tell the Human:
> "agentflow-init.md not found. Ask the Web Reviewer to create it first."

Read the file and extract these parameters:
- `title`
- `branch`
- `worktree`
- `root` (default: `develop` if not specified)

### 2. Verify you are on the root branch

Check the current branch:
```bash
git rev-parse --abbrev-ref HEAD
```

If you are not on the root branch (usually `develop`), stop and tell the Human which branch you are on.

### 3. Run agentflow init

Run the following command with the parameters extracted from `agentflow-init.md`:

```bash
agentflow init \
  --root <root> \
  --branch <branch> \
  --worktree <worktree> \
  --title "<title>"
```

The `--push` flag is enabled by default. If there is no remote configured yet, `agentflow init` will fail with a clear message — tell the Human to add a remote first:
```bash
git remote add origin git@github.com:<user>/<repo>.git
```

### 4. Verify the result

Run:
```bash
agentflow status
```

Confirm that:
- The exchange folder was created
- `phase` is `intake`
- `turn` is `human`
- `next_action` is `run_initial_cli_skill`

### 5. Remove agentflow-init.md

The file has been consumed. Remove it and commit:

```bash
git rm agentflow-init.md
git -C <worktree> add -A
git -C <worktree> commit -m "chore: remove agentflow-init.md after workspace setup"
git -C <worktree> push
```

### 6. Print summary

Print a clear summary:

```
Branch   : <branch>
Worktree : <worktree>
Exchange : <exchange folder path>
Status   : initialized

Next step: Tell the Web Reviewer the workspace is ready on branch <branch>.
           They can now read the exchange folder on GitHub and begin Phase 1 — Discovery.
```

### 7. Stop

Do not begin discovery or any other task. The next turn belongs to the Web Reviewer.
