# Schema

## Content Pack Manifest

Each content pack starts with a `pack.json`.

```json
{
  "id": "arthurian-core",
  "name": "Arthurian Core",
  "myth_region": "Arthurian",
  "tags": ["arthurian", "sword-test"],
  "rooms_file": "rooms.json",
  "entry_room": "stone-yard",
  "story_file": "story_arcs.json",
  "interactions": [
    {
      "id": "example-interaction",
      "when_tags": ["greek", "oracle-network"],
      "adds_tags": ["oracle-dreams-of-avalon"],
      "description": "A short original interaction description."
    }
  ]
}
```

Fields:

- `id`: stable pack identifier.
- `name`: player-facing pack name.
- `myth_region`: broad region label.
- `tags`: world tags seeded by the pack.
- `rooms_file`: relative path to the room document.
- `entry_room`: optional room id for `travel <pack-id>`.
- `story_file`: optional story arc document.
- `interactions`: optional cross-pack reactions to tag combinations.

## Room Schema

`rooms.json` contains a `rooms` array.

```json
{
  "rooms": [
    {
      "id": "stone-yard",
      "name": "Stone Yard",
      "description": "The old test stone sits behind temporary rails.",
      "exits": {
        "north": "round-table-threshold"
      },
      "gated_exits": {
        "east": {
          "target": "avalon-shore",
          "requires_item": "map-scrap",
          "locked_text": "The wall only opens with the right map."
        }
      },
      "encounter": {
        "name": "Seat Claimant",
        "dc": 14,
        "modifier": 1,
        "roll_mode": "advantage",
        "win": "Win text.",
        "lose": "Lose text."
      },
      "recruitable": {
        "name": "Ledger Squire",
        "dc": 11,
        "modifier": 1,
        "roll_mode": "normal",
        "success": "Recruit success text.",
        "failure": "Recruit failure text."
      },
      "item": {
        "id": "map-scrap",
        "name": "Avalon Map Scrap",
        "description": "A short item description."
      },
      "quest": {
        "start": "Quest start text.",
        "start_variants": ["Variant A", "Variant B"],
        "incomplete": "Quest incomplete text.",
        "complete": "Quest complete text."
      },
      "atmosphere": {
        "palette": "rain-gold",
        "weather": "iron rain",
        "myth_layer": "arthurian court",
        "motifs": ["lanterns", "wet banners"]
      }
    }
  ]
}
```

Fields:

- `id`, `name`, `description`: required room identity.
- `exits`: visible exits to other room ids.
- `gated_exits`: exits that require an item.
- `encounter`: optional d20 challenge for `fight`.
- `recruitable`: optional d20 companion for `recruit`.
- `item`: optional visible item for `take`.
- `quest`: optional starter quest state.
- `atmosphere`: room metadata for the visual and audio client.

## Story Arc Schema

`story_arcs.json` contains an `arcs` array.

```json
{
  "arcs": [
    {
      "id": "sword-test",
      "title": "The Sword Test Is Not A Receipt",
      "kind": "main",
      "lore_beats": ["Arthur's kingship is still argued over after the wonder."],
      "source_ids": ["malory-1251", "geoffrey-37848"],
      "summary": "Short summary.",
      "original_hook": "Original Eldermere hook.",
      "required_tags": ["arthurian"],
      "required_factions": {
        "Round Table": 1
      },
      "adds_tags": ["sword-test"],
      "variation_tags": ["stone-version", "lake-version"],
      "steps": [
        {
          "id": "collect-witness-marks",
          "title": "Collect Witness Marks",
          "room_hint": "stone-yard",
          "objective": "Find three incompatible accounts.",
          "commands": ["look", "quest"],
          "outcome_tags": ["witness-contradiction"],
          "faction_effects": {
            "Round Table": 1
          }
        }
      ]
    }
  ]
}
```

Fields:

- `id`: stable arc id for saves and validation.
- `title`: player-facing arc title.
- `kind`: `main` or `side`.
- `lore_beats`: source-grounded beats the arc covers.
- `source_ids`: source ids used for validation.
- `summary`: concise purpose.
- `original_hook`: the original Eldermere angle.
- `required_tags`: tags needed before the arc becomes eligible.
- `required_factions`: minimum faction reputation values.
- `adds_tags`: tags added when the arc completes.
- `variation_tags`: optional branch tags chosen when the arc starts.
- `steps`: ordered objectives.

Each step can include:

- `id`, `title`, `objective`: required fields.
- `room_hint`: room id required for `story next`.
- `commands`: suggested commands for that step.
- `outcome_tags`: tags added when the step advances.
- `faction_effects`: reputation changes applied at that step.

## Player State Schema

The server persists this state per player:

```json
{
  "room_id": "lantern-yard",
  "party": ["Ledger Squire"],
  "items": [
    {
      "id": "excalibur-fragment",
      "name": "Excalibur Fragment",
      "description": "a moon-bright shard of Excalibur"
    }
  ],
  "quest": {
    "Started": true,
    "Completed": false,
    "Variant": "..."
  },
  "story": {
    "active_arc_id": "sword-test",
    "step_index": 0,
    "completed_arc_ids": ["merlins-ledger"],
    "tags": ["arthurian", "sword-test"],
    "variant_tag": "stone-version"
  },
  "factions": {
    "Round Table": 1
  }
}
```

The `quest` object currently serializes with capitalized keys (`Started`, `Completed`, `Variant`) because that struct uses Go's default JSON field naming. The `story` object uses explicit snake_case tags.

## Probability Fields

Room encounters and recruitables share the same d20 probability model.

- `dc`: target number from 2 to 30.
- `modifier`: flat bonus or penalty from room content.
- `roll_mode`: empty, `normal`, `advantage`, or `disadvantage`.
- `crit_win`: optional natural 20 text.
- `crit_lose`: optional natural 1 text.

The runtime also adds:

- combat base bonus: `+2`
- recruitment base bonus: `+1`
- party bonus for combat: up to `+3`

Players can inspect current probabilities with `odds`, `odds fight`, `odds recruit`, and `odds story <id>`.

## HTTP And WebSocket Shapes

- `POST /api/v1/sessions` accepts `{ "display_name": "Wanderer" }` and returns `player_id`, `display_name`, and `token`.
- `GET /ws?player_id=...&token=...` opens the command socket.
- WebSocket command messages can be raw strings or JSON `{ "command": "look" }`.
- WebSocket events use `{ "type": "...", "text": "...", "room": {...} }`.

The `room` payload is what drives the browser room view.
