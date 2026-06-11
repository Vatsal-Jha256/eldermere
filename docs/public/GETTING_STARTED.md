# Getting Started

## Prerequisites

- Docker
- Node.js 20+ and npm, for local web development
- Go 1.23+, optional locally because Docker can run backend checks

## Run The Project

From the repository root:

```sh
docker compose up --build
```

Open:

- Web client: <http://localhost:5173>
- API health: <http://localhost:8080/healthz>
- WebSocket command endpoint: `ws://localhost:8080/ws`
- Session endpoint: `POST http://localhost:8080/api/v1/sessions`

Postgres is exposed on `localhost:5433`. Inside Docker, services still use `db:5432`.

The web client creates a session through `POST /api/v1/sessions`, stores it in browser `localStorage`, and sends the player id plus token to `/ws`. The server persists room location, inventory, party, and quest state in PostgreSQL.

## Run Checks

Backend checks through Docker:

```sh
docker run --rm -v "$PWD/apps/server:/src" -w /src golang:1.26-alpine go test ./...
```

Frontend checks:

```sh
cd apps/web
npm install
npm run check
```

## Current Scaffold

- `apps/server`: Go API server with `/healthz`, `/api/v1/status`, and `/ws`.
- `apps/web`: SvelteKit browser client with a live WebSocket command console.
- `apps/server/internal/game/content/starter/rooms.json`: starter room data loaded by the server.
- `apps/server/internal/storage`: PostgreSQL and in-memory persistence implementations.
- `docker-compose.yml`: Postgres, server, and web services.
- `docs/public`: public Docsify documentation.
- `private-docs`: local learning notes, ignored by git.

## Starter Commands

- `quest`: start or check the starter quest.
- `look`: inspect the current room.
- `go north`, `go east`, `go south`, `go west`: move through room exits.
- `fight`: resolve the current room's encounter with a d20-style roll.
- `recruit`: attempt to recruit the current room's companion.
- `take`: pick up the current room's visible item.
- `inventory`: list carried items.
- `party`: list recruited companions.
- `say hello`: send a local speech event.

Players in the same room receive presence, `say`, fight, and recruit events. New arrivals receive the recent room event log.

## Starter Quest Path

The current vertical slice has a small Arthurian quest arc:

1. Start in Lantern Yard and run `quest`.
2. Explore east into Market Under and try `recruit`.
3. Go `down` to Smuggler Vault.
4. Use `take` to collect the Excalibur Fragment.
5. Return to Lantern Yard and run `quest` again to complete the arc.

Reconnect after picking up the fragment to verify persistence: the same browser should resume in the last room with the item still in inventory. Clearing `localStorage` starts a new session.
