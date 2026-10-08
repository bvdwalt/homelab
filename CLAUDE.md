# Workflow

Always work in a git worktree. At the start of any session that will make changes, call EnterWorktree before touching any files. Each worktree should cover one topic so that commits stay focused and unrelated changes don't get mixed together.

`main` is protected by a ruleset: direct pushes are blocked. When done, push the branch, open a PR, wait for CI, then squash or rebase merge (linear history is required) and delete the branch:

```bash
git push -u origin <branch>
gh pr create --draft
gh pr merge --squash --delete-branch
```
