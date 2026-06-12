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
- Optional `entry_room` points `travel <pack-id>` to a specific room; otherwise the first room is used.
- Pack interactions have ids, trigger tags, and descriptions.
- Optional `story_file` documents have at least one story arc.
- Story arcs include ids, titles, `main` or `side` kind, lore beats, source ids, summaries, original hooks, and steps.
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
  "entry_room": "oracle-jetty",
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
      "atmosphere": {
        "palette": "rain-gold",
        "weather": "iron rain",
        "myth_layer": "arthurian court",
        "motifs": ["lanterns", "river-stone", "wet banners"]
      },
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

## Story Arc Shape

A pack can also include `story_file` in `pack.json`. The server loads validated story arcs at startup and exposes them through `story`, `story start <id>`, `story status`, `story next`, and `story tags`. Required tags lock arcs until earlier story outcomes add the right tags. Steps with `room_hint` require the player to be in that room before `story next` advances. `story start`, `story status`, and `story next` show room hints and suggested commands when provided. Story steps can also change faction reputation through `faction_effects`.

```json
{
  "arcs": [
    {
      "id": "sword-test",
      "title": "The Sword Test Is Not A Receipt",
      "kind": "main",
      "lore_beats": [
        "Arthur's kingship is proved through the sword test, but public acceptance still depends on witnesses and politics."
      ],
      "source_ids": ["malory-1251", "geoffrey-37848"],
      "summary": "Players investigate the sword test as both miracle and political event.",
      "original_hook": "The under-market sells forged witness marks.",
      "required_tags": ["arthurian"],
      "adds_tags": ["sword-test", "contested-kingship"],
      "variation_tags": ["stone-version", "bribed-witness"],
      "steps": [
        {
          "id": "collect-witness-marks",
          "title": "Collect Witness Marks",
          "room_hint": "stone-yard",
          "objective": "Find three incompatible accounts of the sword test.",
          "commands": ["look", "quest"],
          "outcome_tags": ["witness-contradiction"],
          "faction_effects": {
            "Round Table": 1,
            "Camelot Underbelly": -1
          }
        }
      ]
    }
  ]
}
```

Story fields:

- `id`: stable arc id for saves, validation, and future runtime loading.
- `title`: player-facing title.
- `kind`: `main` or `side`.
- `lore_beats`: source-grounded lore beats covered by the arc.
- `source_ids`: ids from `lore/arthurian/sources/SOURCES.md` or a pack's own cited source manifest.
- `summary`: concise arc purpose.
- `original_hook`: original Eldermere story angle built from the lore beats.
- `required_tags`: world tags that must exist before the arc is eligible.
- `adds_tags`: world tags the arc can add for later quests or cross-pack interactions.
- `variation_tags`: branch tags for probabilistic or source-variant outcomes.
- `steps`: playable objectives with optional room hints, commands, outcome tags, and faction effects.

## Writing Rules

- Write original prose.
- Public-domain Arthurian names and motifs are allowed.
- Do not copy dialogue, scenes, character designs, or newly invented details from modern adaptations.
- Prefer hooks that can connect to other legend packs later: faction tags, strange relics, rumors, debts, curses, roads, dreams, and messengers.

## Atmospheric Backgrounds

Rooms can include `atmosphere` metadata. The web client uses this to generate a lightweight atmospheric background, so mods can feel visually distinct without shipping art assets.

Fields:

- `palette`: a named color palette. Existing examples include `rain-gold`, `blackwater`, `candle-smoke`, `tavern-red`, `avalon-green`, `relic-vault`, `coin-shadow`, `oracle-blue`, and `bronze-ash`.
- `weather`: short sensory weather or air description.
- `myth_layer`: the room's mythic context, such as `arthurian court`, `under-market`, or `greek underworld`.
- `motifs`: inspectable visual motifs that should influence future generated art prompts.

The current implementation generates CSS backgrounds. A later renderer can use the same fields as prompts for generated bitmap backgrounds.

## Current Example

See:

- `content-packs/arthurian-core`
- `content-packs/camelot-underbelly`
- `content-packs/greek-crossing`

The Arthurian Core pack demonstrates source-grounded story arcs. The Greek Crossing pack demonstrates the connected-legend rule: it declares interactions that respond to Arthurian tags such as `excalibur-rumor` and `grail-curse` instead of behaving like an isolated Greek zone.

At runtime, content-pack rooms are merged into the world and can be entered with `travel <pack-id>` using the pack's `entry_room`.
