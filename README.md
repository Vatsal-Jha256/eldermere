# Eldermere

Working title for a browser-based MUD and creature-RPG inspired by Arthurian legend, old-school text worlds, tabletop-style probability, and modern web play.

The project direction is documented in [docs/public/PROJECT_PLAN.md](docs/public/PROJECT_PLAN.md).

## Current Status

Stage 0 scaffold is in place. The current repo has:

- Text-first browser MUD interface
- Go backend
- SvelteKit frontend
- WebSocket command loop with movement, speech, fight, recruit, and party commands
- Session-authenticated persistent player state for room, inventory, party, and quest progress
- Room presence, local chat, and recent room event log
- Arthurian starter region
- Runtime-loaded Arthurian story arcs from content packs, with start/status/advance and eligibility commands
- Runtime-loaded content-pack rooms with `travel <pack-id>` entry points
- Recruitable companions, relics, and allies
- Dice/probability-driven encounters with modifiers, advantage/disadvantage, and critical outcomes
- Public modding docs
- Private learning docs for architecture and CSE concepts

The next implementation target is deeper branching from richer room state and broader side-arc playthrough coverage.

## Run Locally

Docker is the recommended path:

```sh
docker compose up --build
```

Then open:

- Web client: <http://localhost:5173>
- API health: <http://localhost:8080/healthz>
- WebSocket command endpoint: `ws://localhost:8080/ws`
- Session endpoint: `POST http://localhost:8080/api/v1/sessions`

Postgres is exposed on `localhost:5433` to avoid conflicts with local Postgres installs on `5432`.

The browser stores an `eldermere.session` object in `localStorage`, created through `POST /api/v1/sessions`. The session token is required by the WebSocket endpoint so location, inventory, party, and quest progress survive reconnects without exposing unauthenticated state changes.

Try `story` in the command console to list loaded Arthurian story arcs, `story eligible` to see playable arcs, `story locked` to see tag/faction locks, `story sword-test` to inspect one, `story start sword-test` to begin, `story next` to advance, and `story tags` to inspect branch tags.

Try `travel arthurian-core`, `travel camelot-underbelly`, or `travel greek-crossing` to enter loaded content-pack regions.

## Checks

```sh
docker run --rm -v "$PWD/apps/server:/src" -w /src golang:1.26-alpine go test ./...
cd apps/web && npm run check
```

## Reference Projects

The main references are:

- [TalesMUD](https://github.com/TalesMUD/talesmud): modern browser MUD direction, Go/Svelte inspiration, WebSocket-first play.
- [Evennia](https://www.evennia.com/): content-first MUD architecture, command/world modeling, builder-friendly design.

This project should learn from those projects without cloning their code, data model, UI, writing, or brand.
