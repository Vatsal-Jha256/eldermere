# Getting Started

## Prerequisites

- Docker and Docker Compose
- Node.js 20+ and npm, for local web development
- Go 1.23+, if you want to run the server directly instead of through Docker

## Run The Game

From the repository root:

```sh
docker compose up --build
```

Open:

- Web client: <http://localhost:5173>
- API health: <http://localhost:8080/healthz>
- Docs: run `make docs-public`, then open <http://localhost:3000>

Postgres is exposed on `localhost:5433`. Inside Docker, services still use `db:5432`.

The web client creates a session through `POST /api/v1/sessions`, stores it in browser `localStorage`, and sends the player id plus token to `/ws`. The server persists room location, inventory, party, quest state, story state, and faction reputation in PostgreSQL.

## Run Checks

Repository checks:

```sh
make test
make validate-content
```

Frontend checks:

```sh
cd apps/web
npm install
npm run check
```

Docs locally:

```sh
make docs-public
```

## Local Dev Loop

If you want split terminals instead of Docker:

```sh
make server
make web
```

`make server` starts the Go API against the local content packs. `make web` starts the browser client on port 5173.

## First Test Path

The smallest end-to-end loop is:

1. Start in Lantern Yard.
2. Run `quest`.
3. Go `west` into Tavern Backroom.
4. Run `take` to collect the Under-Market Map.
5. Return east to Lantern Yard.
6. Run `map` to see the hidden under-route.
7. Run `go under` to enter Smuggler Vault.
8. Run `take` to collect the Excalibur Fragment.
9. Return to Lantern Yard and run `quest` again.

Reconnect after picking up the fragment to verify persistence. The same browser should resume in the last room with the item still in inventory. Clearing `localStorage` starts a new session.

## Environment

The server reads:

- `APP_ENV`
- `SERVER_ADDR`
- `DATABASE_URL`
- `CONTENT_PACKS_DIR`
- `LOG_LEVEL`

The web client reads `PUBLIC_API_BASE` at build time. If it is unset, the client falls back to the current origin, which is the cleanest path for a same-origin reverse proxy deployment.
