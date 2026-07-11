# Agent Instructions

These instructions apply to the whole repository.

## Public-Repo Standard

- Keep public docs concise, professional, and contributor-first.
- Do not add generated-content disclaimers, tool references, hidden prompt notes, or internal planning language.
- Keep plans, scratch notes, learning notes, and release checklists local-only unless they are intentionally rewritten as public contributor docs.
- Do not copy protected modern adaptations, project branding, prose, scenes, or assets.
- Credit inspirations through short acknowledgments, not long internal audits.

## Commits And Pushes

- Use short, direct commit messages, for example `feat: add room validation` or `docs: simplify modding guide`.
- Commit only coherent checkpoints that build or validate.
- Run relevant checks before pushing:
  - Go/backend changes: `cd apps/server && go test ./...`
  - Content changes: `make validate-content`
  - Web changes: `cd apps/web && npm run check`
- Push after a verified checkpoint when the branch is intended to be shared.
- Do not force-push unless the maintainer explicitly asks for history cleanup.

## Contribution Mindset

- Favor small, reviewable changes.
- Keep modding and contribution paths simple.
- Prefer data-driven content over hardcoded one-off behavior.
- Update docs when behavior, commands, schemas, or setup steps change.
- Preserve existing user work; do not revert unrelated changes.
