# Contributing

Eldermere is early. The best contributions are focused, tested, and easy to review.

## Good First Areas

- Add or improve Arthurian lore notes.
- Add rooms, story arcs, relics, companions, or factions.
- Improve content validation.
- Improve command help, accessibility, or mobile layout.
- Tighten documentation when something is confusing.
- Add or improve the procedural atmosphere or audio descriptions in docs.

## Easy Contribution Paths

If you want a small first change, pick one of these:

1. Add a room and validate it with the content toolchain.
2. Add a short story step with a clear `room_hint`.
3. Improve a docs page with a missing command, schema field, or deployment note.
4. Add another Arthurian lore note tied to a source id.
5. Update a room's atmosphere metadata and verify the browser rendering still feels coherent.

For implementation details, read [Gameplay](GAMEPLAY.md), [Schema](SCHEMA.md), [Procedural Systems](PROCEDURAL.md), and [Modding](MODDING.md).

## Content Rules

- Write original prose.
- Use public-domain myth and legend material as source grounding.
- Do not copy dialogue, scenes, character designs, or unique plot inventions from modern adaptations.
- Cite source ids when adding Arthurian story arcs.
- Keep content packs connected through tags, factions, routes, relics, or shared consequences.

## Before A PR

Run:

```sh
make test
make validate-content
```

Keep PRs small when possible. Explain what changed, why it matters, and how you checked it.
