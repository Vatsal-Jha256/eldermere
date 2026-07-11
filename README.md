# Eldermere

Eldermere is an early browser MUD and creature-RPG for exploring Arthurian legend through play.

It mixes room-based exploration, source-grounded story arcs, recruitable companions, relics, faction reputation, and d20-style checks. The project is intentionally mod-friendly: most world content lives in JSON content packs.

The project began as a way to turn interest in Arthurian adaptations into a more source-grounded, playable way to learn the legends. It is currently a prototype and will grow gradually. Contributions are welcome from anyone interested in lore, world building, mechanics, accessibility, or polish.

## Run

```sh
docker compose up --build
```

Then open:

- Web client: <http://localhost:5173>
- API health: <http://localhost:8080/healthz>

## Useful Commands

In the game:

- `help`
- `look`
- `go north`
- `quest`
- `story`
- `story eligible`
- `story start sword-test`
- `fight`
- `recruit`
- `inventory`
- `party`
- `travel arthurian-core`

For development:

```sh
make test
make validate-content
```

## Project Layout

- `apps/server`: Go API, WebSocket command loop, game engine, persistence.
- `apps/web`: SvelteKit browser client.
- `content-packs`: JSON world packs.
- `lore/arthurian`: public-domain source notes and local lore index.
- `docs/public`: Docsify documentation.

## Docs

Run the public docs locally:

```sh
make docs-public
```

Then open <http://localhost:3000>.

## Contributing

Eldermere is nascent. Good contributions are small, tested, and easy to review.

Useful areas:

- Arthurian lore coverage
- Room and quest writing
- Content-pack validation
- DnD-style mechanics
- Browser MUD usability
- Accessibility and responsive UI
- Modding documentation

Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.
