# AgentFlow Init

Read the `agentflow-init-<branch-slug>.md` file created by the Web Reviewer and initialize the AgentFlow workspace.

## Idempotency

Before executing any step, check if it was already done. Skip completed steps silently and continue from where things left off. Never duplicate work or overwrite existing results. If everything is already done, say so and stop.

## Steps

### 1. Pull latest changes from origin

The Web Reviewer created `agentflow-init-<branch-slug>.md` directly on GitHub. Pull to get it locally:

```bash
git pull --rebase
```

### 2. Find and read the agentflow-init file

Look for `agentflow-init-*.md` files in the repository root (current directory, on the root branch).

- If no file is found after pulling, stop and tell the Human:
  > "No agentflow-init-*.md file found after git pull. Ask the Web Reviewer to confirm they committed it to the develop branch."
- If exactly one file is found, use it.
- If multiple files are found, list them and ask the Human which feature to initialize.

Read the file and extract these parameters:
- `title`
- `branch`
- `worktree`
- `root` (default: `develop` if not specified)

### 3. Verify you are on the root branch

Check the current branch:
```bash
git rev-parse --abbrev-ref HEAD
```

If you are not on the root branch (usually `develop`), stop and tell the Human which branch you are on.

### 4. Remove the agentflow-init file from the root branch BEFORE init

This is critical: the init file must be removed from `develop` before the worktree is created, otherwise the file will be inherited by the new feature branch.

Let `<init-file>` be the filename found in step 2.

Check if it still exists:
- If it exists → remove and commit on the root branch:
  ```bash
  git rm <init-file>
  git commit -m "chore: consume <init-file>"
  git push
  ```
- If it was already removed → skip.

### 5. Run agentflow init

Check if the workspace already exists by running `agentflow status`.
- If the exchange folder for this branch already exists → skip to step 5.
- If it does not exist → run:

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

### 6. Verify the result

Run:
```bash
agentflow status
```

Confirm that:
- The exchange folder was created
- `phase` is `intake`
- `turn` is `web`
- `next_action` is `begin_discovery`

### 7. Print summary

```
Branch   : <branch>
Worktree : <worktree>
Exchange : <exchange folder path>
Status   : initialized

Next step: Tell the Web Reviewer the workspace is ready on branch <branch>.
           They can now read the exchange folder on GitHub and begin Phase 1 — Discovery.
```

### 8. Stop

Do not begin discovery or any other task. The next turn belongs to the Web Reviewer.
