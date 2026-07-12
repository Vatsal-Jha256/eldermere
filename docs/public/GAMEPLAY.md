# Gameplay

## Core Loop

Eldermere is a text-first browser MUD.

1. Read the room with `look`.
2. Move with `exits` and `go <direction>`.
3. Use `quest` and `story` to follow the current arc.
4. Use `fight`, `recruit`, `take`, `inventory`, `party`, and `factions` to change state.
5. Use `who` and `say` to play with other people in the same room.

The browser client shows room atmosphere, recent events, example commands, and current room exits. `story next` only advances when the current step requirements are met, including room hints when the arc needs you in a specific place.

## Command Reference

### Movement

- `look` or `l`: inspect the current room.
- `where` or `room`: show the current room name and stable room id.
- `lore` or `atmosphere`: show the room's myth layer, weather, and motifs.
- `exits`: list visible exits.
- `go north`, `go east`, `go south`, `go west`: move through the world.
- `travel arthurian-core`: enter the entry room for a loaded content pack.
- `map`: inspect hidden or gated routes.

### Story

- `quest`: start or check the starter quest.
- `story`: list loaded story arcs.
- `story eligible`: list currently playable arcs.
- `story locked`: list blocked arcs and why they are blocked.
- `story <id>`: inspect one arc.
- `story start <id>`: begin a loaded arc.
- `story status`: show the active arc and the current step.
- `story next`: advance when the step requirements are satisfied.
- `story tags`: list earned branch tags.

### Social

- `who`: list players in your current room.
- `say <text>` or `talk <text>`: send room-local speech.

### Combat And Companions

- `fight`: resolve the current room encounter.
- `recruit`: attempt to recruit the current room companion.
- `odds`: show exact success chances for available fight and recruit checks.
- `odds story <id>`: show uniform story variation odds for an arc.
- `party`: list recruited companions.

### Inventory And State

- `take`: pick up the current room item.
- `inventory`: list carried items.
- `factions`: list reputation changes.

### Help

- `help`
- `help movement`
- `help story`
- `help combat`
- `help inventory`
- `help social`
- `help world`

## Starter Path

The current starter slice begins in Lantern Yard.

1. Run `quest`.
2. Go `west` to Tavern Backroom.
3. Run `take` to collect the Under-Market Map.
4. Return `east` to Lantern Yard.
5. Run `map` to reveal the hidden under-route.
6. Run `go under` to reach Smuggler Vault.
7. Run `take` to collect the Excalibur Fragment.
8. Return to Lantern Yard and run `quest` again.

That loop tests the core quest path, hidden-route gating, and persistence.

## Probability

The probability model is intentionally simple and visible.

- Combat and recruitment use d20 checks.
- `advantage` rolls two d20s and keeps the higher result.
- `disadvantage` rolls two d20s and keeps the lower result.
- Natural 20 always succeeds.
- Natural 1 always fails.
- `odds` computes the exact success chance for the current room's available checks.
- `fight` and `recruit` results show the roll math, critical result when present, and the applied success chance.

The odds command is useful for tuning content because contributors can see whether a room challenge is forgiving, risky, or punishing before changing the JSON.

## Multiplayer

Players who share a room share room events.

- Presence updates are broadcast when a player enters or leaves a room.
- `say` is heard by other players in the same room.
- `fight`, `recruit`, and party-related events are broadcast to the room.
- `who` asks the server for the current room occupancy.
- `who all` lists occupied room ids and player names.
- `where` shows the stable room id when players need to coordinate or report a bug.
- `help multiplayer` explains room-local play, sessions, and what persists.

For online play, the important rule is simple: room IDs are the stable backend keys, while room names are the player-facing labels. Use the room name in the UI, keep the room id as the canonical internal identifier, and expose the id only when it helps debugging or coordination.

Each browser session has its own player id and token. Room, inventory, party, quest, story, and faction state persist across reconnects for the same browser session. Starting a new character creates a fresh saved session.
