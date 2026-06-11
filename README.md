# Eldermere

Working title for a browser-based MUD and creature-RPG inspired by Arthurian legend, old-school text worlds, tabletop-style probability, and modern web play.

The project direction is documented in [docs/public/PROJECT_PLAN.md](docs/public/PROJECT_PLAN.md).

## Current Status

Stage 0 scaffold is in place. The current repo has:

- Text-first browser MUD interface
- Go backend
- SvelteKit frontend
- WebSocket command loop
- Arthurian starter region
- Recruitable companions, relics, and allies
- Dice/probability-driven encounters
- Public modding docs
- Private learning docs for architecture and CSE concepts

The next implementation target is the command loop: `look`, `go`, and `say`.

## Run Locally

Docker is the recommended path:

```sh
docker compose up --build
```

Then open:

- Web client: <http://localhost:5173>
- API health: <http://localhost:8080/healthz>

Postgres is exposed on `localhost:5433` to avoid conflicts with local Postgres installs on `5432`.

## Checks

```sh
docker run --rm -v "$PWD/apps/server:/src" -w /src golang:1.23-alpine go test ./...
cd apps/web && npm run check
```

## Reference Projects

The main references are:

- [TalesMUD](https://github.com/TalesMUD/talesmud): modern browser MUD direction, Go/Svelte inspiration, WebSocket-first play.
- [Evennia](https://www.evennia.com/): content-first MUD architecture, command/world modeling, builder-friendly design.

This project should learn from those projects without cloning their code, data model, UI, writing, or brand.
