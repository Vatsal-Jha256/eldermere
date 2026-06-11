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

Postgres is exposed on `localhost:5433`. Inside Docker, services still use `db:5432`.

## Run Checks

Backend checks through Docker:

```sh
docker run --rm -v "$PWD/apps/server:/src" -w /src golang:1.23-alpine go test ./...
```

Frontend checks:

```sh
cd apps/web
npm install
npm run check
```

## Current Scaffold

- `apps/server`: Go API server with `/healthz` and `/api/v1/status`.
- `apps/web`: SvelteKit browser client with a playable mock command console.
- `docker-compose.yml`: Postgres, server, and web services.
- `docs/public`: public Docsify documentation.
- `private-docs`: local learning notes, ignored by git.

