# Contributing

Eldermere is early. Contributions should make the game easier to run, easier to mod, or more playable.

## Local Setup

Use Docker first:

```sh
docker compose up --build
```

Services:

- Web client: <http://localhost:5173>
- API health: <http://localhost:8080/healthz>
- Postgres: `localhost:5433`

## Before Opening A Pull Request

Run:

```sh
docker run --rm -v "$PWD/apps/server:/src" -w /src golang:1.26-alpine go test ./...
cd apps/web && npm run check
```

If Go is installed locally, `make test-server` also works.

## Content Contributions

Content should be original writing. Arthurian legend names and public-domain myth material are allowed, but do not copy dialogue, character designs, scenes, or unique plot inventions from modern adaptations.

Future content packs should connect to the shared world model rather than acting like isolated zones.
