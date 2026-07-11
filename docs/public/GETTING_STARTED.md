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

## Project Layout

- `apps/server`: Go API server with `/healthz`, `/api/v1/status`, and `/ws`.
- `apps/web`: SvelteKit browser client with a live WebSocket command console.
- `apps/server/internal/game/content/starter/rooms.json`: starter room data loaded by the server.
- `content-packs`: validated mod packs loaded by local and Docker server runs.
- `apps/server/internal/storage`: PostgreSQL and in-memory persistence implementations.
- `docker-compose.yml`: Postgres, server, and web services.
- `docs/public`: public Docsify documentation.

Room backgrounds are generated from room `atmosphere` metadata: palette, weather, myth layer, and motifs. The browser combines CSS atmosphere layers with a procedural canvas backdrop, and the same metadata also shapes ambient audio.

## Starter Commands

- `help`: list command families and discover focused help topics.
- `help story`, `help movement`, `help combat`, `help inventory`, `help social`, `help world`: inspect focused MUD-style help topics.
- `quest`: start or check the starter quest.
- `story`: list loaded story arcs from content packs.
- `story eligible`: list currently playable story arcs.
- `story locked`: list blocked story arcs and the missing tags or faction reputation.
- `story sword-test`: inspect a source-grounded Arthurian story arc.
- `story start sword-test`: begin a loaded story arc.
- `story status`: inspect active story progress, including the room and suggested commands for the current step.
- `story next`: advance the active story arc when you are in the required room, collecting outcome tags and faction effects.
- `story tags`: inspect earned branch and eligibility tags.
- `factions`: inspect reputation changes from encounters and story steps.
- `travel arthurian-core`: move to a loaded content pack's entry room.
- `look`: inspect the current room.
- `go north`, `go east`, `go south`, `go west`: move through room exits.
- `fight`: resolve the current room's encounter with a d20-style roll, including tuned modifiers, advantage/disadvantage, and critical outcomes where the room defines them.
- `recruit`: attempt to recruit the current room's companion with the same d20 check model.
- `take`: pick up the current room's visible item.
- `inventory`: list carried items.
- `party`: list recruited companions.
- `map`: inspect hidden or gated routes from the current room.
- `say hello` or `talk hello`: send a local speech event.

Players in the same room receive presence, `say`, fight, and recruit events. New arrivals receive the recent room event log.

## Starter Quest Path

The current vertical slice has a small Arthurian quest arc:

1. Start in Lantern Yard and run `quest`.
2. Go `west` into Tavern Backroom and use `take` to collect the Under-Market Map.
3. Return east to Lantern Yard, then use `map` to see the unlocked under-route.
4. Use `go under` to reach Smuggler Vault.
5. Use `take` to collect the Excalibur Fragment.
6. Return to Lantern Yard and run `quest` again to complete the arc.

Reconnect after picking up the fragment to verify persistence: the same browser should resume in the last room with the item still in inventory. Clearing `localStorage` starts a new session.
