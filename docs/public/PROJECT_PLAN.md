# Project Plan

## Summary

Mythweald is an open-source browser MUD and creature-RPG. It starts in Arthurian legend, then grows into a connected myth universe where legend packs interact through shared factions, travel, prophecy, reputation, curses, relics, and world events.

The game should feel text-first and fast to build: room prose, command input, static backgrounds, compact character panels, and probabilistic encounters. Visual maps can arrive later, and only when the player has an in-world map, relic, guide, or equivalent reason to see one.

## Product Direction

- Start with Arthurian legend as the core region.
- Use real public-domain legend names where useful: Arthur, Merlin, Camelot, Avalon, Excalibur, Morgan, Mordred, the Grail, and the Round Table.
- Do not copy protected modern adaptations: no borrowed dialogue, scene structure, plot inventions, visual designs, branding, or exact characterizations from films, games, books, or TV.
- Use modern crime-caper pacing, banter, debts, factions, and betrayals as tonal inspiration, not as copied material.
- Let tone vary by scene: street-myth energy for normal play, darker mythic weight when stakes rise, and occasional weird/cozy side content when it serves the world.

## Core Gameplay

- Players explore rooms through commands such as `look`, `go`, `talk`, `fight`, `recruit`, `equip`, `use`, and `quest`.
- Encounters use tabletop-like probabilities: stats, dice rolls, advantage/disadvantage, critical outcomes, resistances, and risky bargains.
- Collection combines three categories:
  - Companions: beasts, spirits, squires, cursed allies, local legends, and mythic creatures.
  - Relics: blessings, curses, techniques, fragments, oaths, charms, and named items.
  - Allies: knights, witches, outlaws, rivals, mercenaries, priests, scholars, and faction agents.
- Known legends should not be solved by memory. Use random encounter tables, mutable loyalties, hidden motives, alternate quest branches, and probability-driven events.

## Connected Legend Universe

Later legends should not be separate theme parks. They should connect to the same world model.

- Each legend pack adds regions, factions, characters, companions, relics, events, and quest chains.
- Packs can reference and affect each other through shared tags, faction relations, prophecy keys, relic ownership, travel routes, and world events.
- Cross-legend play should be earned through story or systems: ships, roads, portals, dreams, underworld routes, divine messengers, cursed maps, or political invitations.
- A Greek pack, for example, should be able to react to Arthurian state: Excalibur rumors can affect Olympus politics, a Grail curse can disturb an underworld route, or a Round Table faction can hire a Greek seer.
- The engine should support cross-pack interactions without hardcoding every pair. Content packs should declare relationships through data, and the server should resolve eligible events.

## Technical Direction

- Backend: Go.
- Frontend: SvelteKit.
- Realtime: WebSockets.
- Database: PostgreSQL.
- Later cache/queue option: Redis only when needed.
- Deployment: Docker and Docker Compose first.
- Docs: Docsify.

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

## Milestones

1. Planning and repository setup.
2. Go server skeleton with health route, config, logging, and test setup.
3. SvelteKit client skeleton with playable first screen layout.
4. WebSocket command loop with `look`, `go`, and `say`.
5. Room/content model loaded from data files.
6. Arthurian starter region with 5-8 rooms and one short quest.
7. Dice engine, encounter engine, and one recruitable companion/relic/ally.
8. Persistence for accounts, characters, inventory, location, and quest state.
9. Public modding guide and content-pack validation.
10. Private learning docs explaining each major system and checkpoint.

## Open Source And Modding

- Public content packs should be data-first and reviewable.
- Mods should be able to add rooms, NPCs, encounters, companions, relics, factions, and quests.
- Content schemas should prefer clear IDs and tags over brittle code hooks.
- The project should include a small example mod pack as soon as the core loop works.
- Contributions should require tests for engine changes and validation for content changes.
