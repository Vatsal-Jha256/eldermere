# References

## TalesMUD

- Site/repository: <https://github.com/TalesMUD/talesmud>
- Project docs page found during planning: <https://github.com/TalesMUD/talesmud/blob/master/PROJECT.md>

Relevant ideas:

- Browser-based MUDs can feel modern without abandoning text-first play.
- Go and Svelte are a credible stack for realtime MUD development.
- WebSockets, persistent state, graphical overlays, mobile-responsive UI, and content tooling are useful long-term targets.

How Eldermere should use it:

- Treat TalesMUD as a product and architecture reference.
- Do not copy implementation details, UI, world content, writing, or branding.
- Re-check license and code boundaries before adopting any code directly.

## Evennia

- Site: <https://www.evennia.com/>
- GitHub: <https://github.com/evennia/evennia>
- Introduction: <https://www.evennia.com/docs/latest/Evennia-Introduction.html>

Relevant ideas:

- MUD engines benefit from being game-agnostic at the lower layers.
- Rooms, characters, objects, commands, channels, persistence, and builder workflows are core concepts.
- A strong engine should avoid prescribing combat rules, classes, races, or genre too early.

How Eldermere should use it:

- Treat Evennia as a design reference for flexible MUD architecture.
- Study its concepts, not its exact Python/Django/Twisted implementation.
- Keep game-specific systems such as dice combat, companions, relics, and legend packs separate from the lower-level command/world engine.

## Public Domain And Adaptations

Arthurian legends are old public-domain source material in broad terms, but modern adaptations can add protected expression.

Use:

- Historical and medieval Arthurian names and motifs.
- Original writing, mechanics, factions, quests, and art direction.
- Broad genre influences such as fast banter, heists, debts, betrayals, and underworld politics.

Avoid:

- Copying modern film/game/book dialogue.
- Copying unique scene sequences from modern adaptations.
- Copying specific visual designs, branding, logos, posters, UI, or character styling.
- Marketing the game as connected to any protected adaptation.
