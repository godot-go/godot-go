---
name: opening-pull-requests
description: Use when opening or updating a GitHub pull request in this repository — titling a PR, composing its body from OpenSpec artifacts, choosing the base branch for stacked work, linking an issue, or force-pushing a rebased stack branch. Supersedes generic PR-body templates for this repo.
metadata:
  author: godot-go
---

# Opening Pull Requests (godot-go)

One OpenSpec change = one branch = one PR. The PR body is never written fresh — it is the change's own planning artifacts. The title is not composed — it is copied.

## The contract

- **Branch:** `feat-<change-name>` / `fix-<change-name>`, off `main` (or off the branch below it in a stack).
- **Commits on the branch:** `docs: propose <change> change` → `feat:`/`fix: <implementation>` → `docs: <doc side>` → `chore: archive <change> change`. Conventional Commits, subject-only.
- **PR title:** the implementation commit's subject, verbatim (e.g. `feat: register editor plugins with session guards and auto-unload`). Do not reword it; do not add `(#NNN)` — squash-merge appends that.
- **PR body:** the change's `proposal.md`, a `---` horizontal-rule separator, then its `design.md` — all verbatim. No extra sections, no summary rewrite, no generic template (Summary/Changes/Test-plan bodies are NOT this repo's convention).
- **Issue link:** if the change resolves a GitHub issue, append one `Closes #NNN` line at the end of the body. Roadmap feature changes usually have no issue and no line.
- **Base:** `main` for solo work; the branch immediately below for stacked PRs.
- **Create and verify:**

```bash
git push -u origin <branch>
CHANGE=openspec/changes/archive/<dated-change-dir>
{ cat $CHANGE/proposal.md; printf '\n---\n\n'; cat $CHANGE/design.md; } > /tmp/opencode/pr-body.md
printf '\nCloses #%s\n' <NNN> >> /tmp/opencode/pr-body.md   # only when the change resolves an issue
gh pr create --base <base> --head <branch> --title "$(git log -1 --format=%s <implementation-commit>)" --body-file /tmp/opencode/pr-body.md
gh pr view --json number,title,baseRefName   # confirm what GitHub recorded
```

The archive commit precedes the PR precisely so the body reads from the archived artifact paths.

## Stacked branches

After lower PRs squash-merge into `main`, replay only the branch's own commits:

```bash
git fetch origin --prune
git rebase --onto origin/main <old-base-tip> <branch>   # plain `git rebase origin/main` DUPLICATES the squashed commits — patch-id does not match a squash
git diff <pre-rebase-sha> HEAD                          # must be empty: no content drift
git push --force-with-lease=<branch>:<known-old-sha> origin <branch>
```

## Merge-order hazard (proven here by #154)

A stacked PR squash-merges into **its base branch**, not `main`. If that base is an intermediate branch later deleted, the squash content is orphaned and `main` silently lacks it. Before merging any PR in a stack, merge bottom-up and ensure each merge either deletes its branch (`gh pr merge --squash --delete-branch` — GitHub then retargets child PRs) or you manually retarget every child first (`gh pr edit <child> --base main`). Recovery if orphaned anyway: the squash commit is still fetchable via `refs/pull/<N>/head` — cherry-pick it onto `main`.

## Common mistakes

| Mistake | Fix |
| --- | --- |
| Paraphrasing or authoring a PR body | Concatenate `proposal.md` + `design.md` |
| Generic template sections | Not this repo's convention |
| Plain `git rebase origin/main` on a stack branch | Use `git rebase --onto <new> <old-tip> <branch>` |
| Bare `git push -f` | `--force-with-lease` pinned to the known old sha |
| Merging a parent PR without branch-deletion or child retarget | Squash orphans on the intermediate branch |
| Title hand-composed differently from commits | Copy the implementation commit subject |
