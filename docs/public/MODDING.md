# Modding

Eldermere content packs are data-first. The current validator supports pack manifests plus room packs in JSON.

## Validate A Pack

From the repository root:

```sh
make validate-content
```

Or validate a specific file:

```sh
cd apps/server
go run ./cmd/eldermere-content validate ../../content-packs/camelot-underbelly
```

The validator checks:

- Pack manifest has `id`, `name`, `myth_region`, and `rooms_file`.
- Pack interactions have ids, trigger tags, and descriptions.
- At least one room exists.
- Every room has an `id`, `name`, and `description`.
- Room ids are unique.
- Every exit points to an existing room.
- Empty exit directions are rejected.

## Room Pack Shape

Each pack has a `pack.json`:

```json
{
  "id": "greek-crossing",
  "name": "Greek Crossing",
  "myth_region": "Greek",
  "tags": ["greek", "oracle-network", "underworld-route"],
  "rooms_file": "rooms.json",
  "interactions": [
    {
      "id": "excalibur-rumor-reaches-oracle",
      "when_tags": ["arthurian", "excalibur-rumor"],
      "adds_tags": ["oracle-seeks-foreign-steel"],
      "description": "An oracle hears of Excalibur and sends a dream-message toward Camelot."
    }
  ]
}
```

And a `rooms.json`:

```json
{
  "rooms": [
    {
      "id": "coin-arch",
      "name": "Coin Arch",
      "description": "A low arch under Camelot...",
      "exits": {
        "east": "candle-court"
      },
      "encounter": {
        "name": "Ledger Knight",
        "dc": 13,
        "win": "Win text.",
        "lose": "Lose text."
      },
      "recruitable": {
        "name": "Candle Page",
        "dc": 10,
        "success": "Recruit success text.",
        "failure": "Recruit failure text."
      },
      "item": {
        "id": "excalibur-fragment",
        "name": "Excalibur Fragment",
        "description": "a visible item description"
      }
    }
  ]
}
```

## Writing Rules

- Write original prose.
- Public-domain Arthurian names and motifs are allowed.
- Do not copy dialogue, scenes, character designs, or newly invented details from modern adaptations.
- Prefer hooks that can connect to other legend packs later: faction tags, strange relics, rumors, debts, curses, roads, dreams, and messengers.

## Current Example

See:

- `content-packs/camelot-underbelly`
- `content-packs/greek-crossing`

The Greek Crossing pack demonstrates the connected-legend rule: it declares interactions that respond to Arthurian tags such as `excalibur-rumor` and `grail-curse` instead of behaving like an isolated Greek zone.
