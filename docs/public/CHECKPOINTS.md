# Checkpoint Workflow

## GitHub Setup

When the GitHub repository is created, add it as `origin` and push the first checkpoint.

Recommended repository settings:

- Public repo when the first public docs are ready.
- MIT or Apache-2.0 license for code, unless a different open-source strategy is chosen.
- Issue labels for `engine`, `client`, `content`, `docs`, `modding`, `learning`, `bug`, and `design`.
- Protect `main` after the first stable scaffold.

## Local Checkpoints

Use small commits that map to real milestones.

Suggested first commits:

1. `docs: add project plan and references`
2. `chore: scaffold go server`
3. `chore: scaffold sveltekit client`
4. `feat: add websocket command loop`
5. `feat: load rooms from content files`
6. `feat: add starter arthurian quest`
7. `docs: add modding guide`

## Branch Style

- `main`: stable, runnable project state.
- `feature/server-scaffold`: Go backend skeleton.
- `feature/client-scaffold`: SvelteKit frontend skeleton.
- `feature/command-loop`: WebSocket and command parser.
- `feature/content-packs`: data loading and validation.
- `docs/learning-notes`: private/local learning docs only if kept outside public GitHub.

## Private Docs Rule

Keep personal learning notes in `private-docs/`, which is ignored by git by default.

If private docs need their own GitHub later, make a separate private repository. Do not publish personal notes inside the open-source game repo unless intentionally reviewed.

