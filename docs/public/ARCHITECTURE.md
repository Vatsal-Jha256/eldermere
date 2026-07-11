# Architecture

## System Overview

Eldermere has three main layers:

1. Browser client in `apps/web`
2. Go server in `apps/server`
3. Data-first world content in `content-packs` and `lore`

The browser handles presentation and input. The Go server handles sessions, room state, story logic, and multiplayer broadcasts. Content packs define the world.

## Runtime Flow

```text
Browser
  -> POST /api/v1/sessions
  -> GET /ws?player_id=...&token=...
Go router
  -> load starter world
  -> merge content packs
  -> validate rooms, arcs, and source refs
  -> serve HTTP + WebSocket
Room hub
  -> track presence and recent room history
Postgres
  -> store player session records and persistent player state
```

### Browser Client

The client in `apps/web`:

- creates a player session
- opens the WebSocket
- renders the current room using room atmosphere metadata
- plays ambient audio cues
- shows command history, example commands, exits, and room status

### Server

The server in `apps/server`:

- exposes `GET /healthz`, `GET /api/v1/status`, `POST /api/v1/sessions`, and `GET /ws`
- loads the starter world from embedded JSON
- loads content packs from `CONTENT_PACKS_DIR` or local defaults
- validates rooms, exits, story arcs, and source ids
- persists player state in PostgreSQL

### Multiplayer Hub

The room hub keeps a per-room client set and recent room history.

- room joins broadcast presence
- room leaves broadcast presence
- new arrivals receive recent room events
- `who` asks the hub for the current room occupancy

This design keeps multiplayer simple: room ids are the backend truth, and room membership is what drives shared events.

### Storage

The PostgreSQL store holds two tables:

- `player_accounts`: player id, display name, and token hash
- `player_states`: room, items, party, faction reputation, and story state

The in-memory store exists for tests and local development; the Postgres store is the normal online path.

## Content Loading

Startup order:

1. Load the embedded starter world.
2. Load content packs from disk.
3. Merge pack rooms, story arcs, and entry room mappings.
4. Validate runtime references against the merged world and Arthurian source ids.

This means a pack can declare a room that lives elsewhere in the merged world, but broken exits or story hints fail validation early.

## Configuration

The server reads:

- `APP_ENV`
- `SERVER_ADDR`
- `DATABASE_URL`
- `CONTENT_PACKS_DIR`
- `LOG_LEVEL`

The browser reads `PUBLIC_API_BASE` from the web runtime environment.

If you host the web client and API on the same origin behind a reverse proxy, `PUBLIC_API_BASE` can be omitted and the browser will fall back to the current origin.

## Deployment Shape

The current repo supports two practical online shapes:

- Separate origins: web client points at a public API URL.
- Same origin: reverse proxy routes `/` to the web client and `/api` plus `/ws` to the Go server.

The same-origin shape is cleaner for production because it avoids cross-origin session and WebSocket wiring in the browser.
