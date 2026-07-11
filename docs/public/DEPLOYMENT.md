# Deployment

## Local

The fastest local run is still:

```sh
docker compose up --build
```

That gives you Postgres, the Go server, and the browser client.

## Free Docs Hosting

The docs are static. They can be deployed separately from the game.

Good free options:

- GitHub Pages for public repositories
- Cloudflare Pages
- Netlify

`docs/public` is the static site root. The docs do not need the Go server to render.

## Free Game Hosting

There are three practical free-hosting paths for this project.

### Render

Render is the most straightforward free-tier option for this repo because it supports web services, static sites, and PostgreSQL in one place.

Use it when:

- you want the least setup friction
- you want a separate static docs host
- you want to ship a short public playtest quickly

Tradeoffs:

- Render’s free services are not meant for production workloads.
- Free Render Postgres databases have a fixed 1 GB storage cap and expire after 30 days.
- Free compute is suitable for previews and hobby use, not a long-running public shard.

For this game, Render is the best “get it online fast” option.

### Koyeb

Koyeb is a good container-first option if you want a more global deployment model, but the free tier is very small.

Use it when:

- you want container deployment
- you want the app available across multiple regions on paid plans later
- you are okay with a very small preview footprint for now

Tradeoffs:

- The free instance is limited to one per organization.
- The free instance is only 0.1 vCPU, 512 MB RAM, and 2 GB SSD.
- Koyeb’s own docs describe the free instance as a preview / hobby tier, not production.

For Eldermere, Koyeb is workable for a tiny demo, but it is tighter than Render for a WebSocket-heavy MUD with a browser client and Postgres.

### Tunnel For Short Playtests

If you only need a temporary test with friends, a tunnel is still the simplest path.

- Cloudflare Quick Tunnels are intended for testing and development, not production.
- They are useful when you do not want to expose a permanent public service yet.

## Recommendation

For this repo:

1. Use **Render** for the fastest public playtest.
2. Use **Koyeb** if you prefer container-first deployment and can live with a smaller free footprint.
3. Use a paid or always-on host once you want a real persistent public shard.

## Recommended Production Shape

If you want one stable public deployment, use this split:

1. Host the Go server and PostgreSQL on Render, Koyeb, or another always-on backend.
2. Put the browser client on the same origin or a reverse proxy in front of the API.
3. Host the docs on GitHub Pages, Cloudflare Pages, or Netlify.

## Practical Steps

1. Provision the host.
2. Install Docker and Docker Compose if you are running containers yourself.
3. Set `DATABASE_URL`, `SERVER_ADDR`, `CONTENT_PACKS_DIR`, and `APP_ENV=production`.
4. Set `PUBLIC_API_BASE` in the web host environment to the public API URL if the web and API are on different origins.
5. If the web and API share one origin behind a reverse proxy, leave `PUBLIC_API_BASE` unset and route `/api` and `/ws` to the server.
6. Deploy `docs/public` to a static host.

## What To Use For Multiplayer

Use the room model already in the game:

- room ids are the stable backend key
- room names are the player-facing label
- shared room membership drives presence and local speech

That is the right model for an online MUD. It keeps the network contract simple and leaves the player UI readable.
