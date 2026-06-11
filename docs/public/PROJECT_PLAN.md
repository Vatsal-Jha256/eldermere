# Arthurian Browser MUD Creature-RPG Plan

## Summary

Build an open-source browser MUD inspired by Arthurian legend, with Pokemon-like collecting, DnD-style probability, and mod-friendly content. The idea is differentiated: MUDs, monster collectors, and Arthurian RPGs all exist, but the specific mix of browser MUD + recruitable companions/relics/knights + moddable legend universe is not obviously saturated.

Arthurian names are broadly usable because the medieval legends are public-domain material, but avoid copying modern adaptations' dialogue, scenes, character designs, or newly invented plot details. Use "Guy Ritchie energy" as tone inspiration, not as source material. References checked: TalesMUD, Evennia, Arthurian copyright discussion, and public-domain Arthur notes.

Eldermere is the current working title. It starts in Arthurian legend, then grows into a connected myth universe where legend packs interact through shared factions, travel, prophecy, reputation, curses, relics, and world events.

The game should feel text-first and fast to build: room prose, command input, static backgrounds, compact character panels, and probabilistic encounters. Visual maps can arrive later, and only when the player has an in-world map, relic, guide, or equivalent reason to see one.

## Key Decisions

- Stack: Go backend, SvelteKit frontend, WebSockets, PostgreSQL, Docker.
- UX: browser-based MUD first, not telnet-only.
- Visual style: text-first rooms with static illustrated backgrounds at launch; map appears only when the player has an in-world map or map-like item.
- Tone: flexible, with street-myth banter in normal play and darker mythic prose during serious arcs.
- Combat/collection: combine all three collection types:
  - Companions: beasts, spirits, cursed allies, squires.
  - Relics/cards: blessings, curses, techniques, Excalibur fragments.
  - Knight/hero units: named allies, rivals, outlaws, witches.
- Docs: both public and private Docsify sites.
  - Public docs: install, modding, content schema, contribution guide.
  - Private docs: learning notes explaining Go concurrency, WebSockets, persistence, game loops, ECS/state modeling, testing, deployment.

## Product Direction And Legal Boundaries

- Start with Arthurian legend as the core region.
- Use real public-domain legend names where useful: Arthur, Merlin, Camelot, Avalon, Excalibur, Morgan, Mordred, the Grail, and the Round Table.
- Do not copy protected modern adaptations: no borrowed dialogue, scene structure, plot inventions, visual designs, branding, or exact characterizations from films, games, books, or TV.
- Use modern crime-caper pacing, banter, debts, factions, and betrayals as tonal inspiration, not as copied material.
- Let tone vary by scene: street-myth energy for normal play, darker mythic weight when stakes rise, and occasional weird/cozy side content when it serves the world.

## Core Gameplay

- Players explore rooms through commands such as `look`, `go`, `talk`, `fight`, `recruit`, `equip`, `use`, and `quest`.
- Encounters use DnD-style probabilities: stats, dice rolls, advantage/disadvantage, critical outcomes, resistances, and risky bargains.
- The first playable version should prove the fun core loop before full multiplayer: create character, enter rooms, move with commands, inspect background/text, recruit one companion/relic/ally, fight one probabilistic battle, and complete one short Arthurian quest arc.
- Collection combines three categories:
  - Companions: beasts, spirits, squires, cursed allies, local legends, and mythic creatures.
  - Relics: blessings, curses, techniques, fragments, oaths, charms, and named items.
  - Allies: knights, witches, outlaws, rivals, mercenaries, priests, scholars, and faction agents.
- Known legends should not be solved by memory. Use random encounter tables, mutable loyalties, hidden motives, alternate quest branches, and probability-driven events.

## Connected Legend Universe

Later legends should not be separate theme parks. Greek, Celtic, Norse, South Asian, and other legend regions can arrive as content packs, but they must connect to the same world model.

- Each legend pack adds regions, factions, characters, companions, relics, events, and quest chains.
- Packs can reference and affect each other through shared tags, faction relations, prophecy keys, relic ownership, travel routes, and world events.
- Cross-legend play should be earned through story or systems: ships, roads, portals, dreams, underworld routes, divine messengers, cursed maps, or political invitations.
- A Greek pack, for example, should be able to react to Arthurian state: Excalibur rumors can affect Olympus politics, a Grail curse can disturb an underworld route, or a Round Table faction can hire a Greek seer.
- The engine should support cross-pack interactions without hardcoding every pair. Content packs should declare relationships through data, and the server should resolve eligible events.

This changes the original Stage 5 from "add Greek, Celtic, Norse, South Asian, or other legend regions as separate content packs without breaking Arthurian v1" to a stronger requirement: those content packs must also connect and interact with each other in the same universe.

## Technical Direction

- Backend: Go.
- Frontend: SvelteKit.
- Realtime: WebSockets.
- Database: PostgreSQL.
- Later cache/queue option: Redis only when needed.
- Deployment: Docker and Docker Compose first.
- Docs: Docsify.
- Public docs: install, modding, content schema, contribution guide, and contributor workflow.
- Private docs: learning notes for Go concurrency, WebSockets, persistence, game loops, ECS/state modeling, testing, and deployment.

The backend should expose a command-driven game loop inspired by MUD architecture: parse command, validate actor state, apply world rules, persist changes, broadcast room/player updates, and return a readable event log.

The frontend should be a browser-native MUD client, not a generic marketing site. First screen should be playable: terminal, room background, room text, exits, inventory/party panel, and event feed.

## Reference Use

TalesMUD is a reference for:

- Browser-first MUD experience.
- Go and Svelte viability.
- WebSocket play.
- Modern terminal plus graphical overlays.
- Content/editor ambition.

Evennia is a reference for:

- Game-agnostic MUD architecture.
- Object, character, room, command, channel, and persistence concepts.
- Builder-friendly content workflows.
- Keeping the engine flexible instead of forcing one combat or class system.

This project should not fork, clone, or copy either reference. The implementation should use original code and its own data model unless a dependency is deliberately adopted under its license.

## Implementation Stages

1. Stage 0: scaffold repo with Go API/server, SvelteKit client, Docker Compose, PostgreSQL, public Docsify docs, and private local Docsify notes.
2. Stage 1: single-player playable core loop with character creation, rooms, movement, room backgrounds/text, one recruitable companion/relic/ally, one probabilistic battle, and one short Arthurian quest arc.
3. Stage 2: browser-MUD multiplayer with accounts/session auth, WebSocket command stream, room presence, local chat, and shared event log.
4. Stage 3: modding system with JSON or YAML content packs for rooms, NPCs, encounters, drops, factions, quests, validation CLI, and an example "Camelot Underbelly" mod pack.
5. Stage 4: richer world with map-gated navigation, factions, party encounters, and procedural quest variations so legend-knowers still get surprises.
6. Stage 5: expanded legend universe where Greek, Celtic, Norse, South Asian, and other legend packs interact with the same world state rather than sitting in separate game modes.

## Original Stage Details To Preserve

Stage 0: scaffold repo with Go API/server, SvelteKit client, Docker Compose, PostgreSQL, Docsify public/private docs.

Stage 1: single-player playable core loop:

- Create character.
- Enter rooms.
- Move with commands.
- Inspect background/text.
- Recruit one companion/relic/ally.
- Fight one probabilistic battle.
- Complete one short Arthurian quest arc.

Stage 2: browser-MUD multiplayer:

- Accounts/session auth.
- WebSocket command stream.
- Room presence.
- Local chat.
- Shared event log.

Stage 3: modding system:

- JSON/YAML content packs for rooms, NPCs, encounters, drops, factions, quests.
- Validation CLI.
- Example "Camelot Underbelly" mod pack.

Stage 4: richer world:

- Map-gated navigation.
- Factions.
- Party encounters.
- Procedural quest variations so legend-knowers still get surprises.

Stage 5: expanded legend universe:

- Add Greek, Celtic, Norse, South Asian, or other legend regions as content packs without breaking Arthurian v1.
- Ensure those legend regions connect and interact through the shared world model.

## Near-Term Milestones

1. Planning and repository setup.
2. Go server skeleton with health route, config, logging, and test setup.
3. SvelteKit client skeleton with playable first screen layout.
4. WebSocket command loop with `look`, `go`, and `say`.
5. Room/content model loaded from data files. Initial starter room loading is in place; the next version should make this mod-pack friendly.
6. Arthurian starter region with 5-8 rooms and one short quest. Initial 6-room "stolen Excalibur fragment" arc is in place.
7. Dice engine, encounter engine, and one recruitable companion/relic/ally. Initial `fight`, `recruit`, and `party` commands are in place for the starter rooms.
8. Persistence for accounts, characters, inventory, location, and quest state.
9. Public modding guide and content-pack validation.
10. Private learning docs explaining each major system and checkpoint.

## Open Source And Modding

- Public content packs should be data-first and reviewable.
- Mods should be able to add rooms, NPCs, encounters, companions, relics, factions, and quests.
- Content schemas should prefer clear IDs and tags over brittle code hooks.
- The project should include a small example mod pack as soon as the core loop works.
- Contributions should require tests for engine changes and validation for content changes.

## Test Plan

- Unit tests for dice/probability, recruit/capture logic, combat resolution, inventory, room movement, and content validation.
- Integration tests for WebSocket commands: connect, look, move, say, fight, and recruit.
- Content tests to ensure every room exit points to a valid room and every quest can complete.
- Basic browser tests for command input, room rendering, background display, and responsive layout.
- Docs checks to ensure the public modding guide includes a working example content pack and private docs explain each implemented system.

## Assumptions

- The first milestone optimizes for fun core loop over full multiplayer.
- Go and SvelteKit are chosen for learning value, portfolio value, and fit with browser MUD architecture, even though Evennia would ship faster.
- The project uses real Arthurian names but original writing, original mechanics, and original visual identity.
- The first public release should be small, moddable, and playable rather than lore-complete.
