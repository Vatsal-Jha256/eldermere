# Procedural Systems

## What Is Procedural

Eldermere is not fully hand-authored and not fully random. It mixes authored world content with deterministic generation around the edges.

Authored:

- rooms
- story arcs
- items
- faction effects
- pack interactions

Generated:

- room backdrops
- room mood layers
- ambient audio beds
- short cue sounds for events
- motion inside the browser atmosphere canvas
- exact d20 odds shown through the `odds` command

The generated pieces are seeded from room metadata, so two players can see and hear the same room in a consistent way without shipping art or audio assets for every location.

## Atmosphere Profile

The browser builds an atmosphere profile from:

- room id
- palette
- weather
- myth layer
- motifs

That profile produces:

- a stable seed
- a detected biome
- a small set of sound modes
- visual generation settings

The implementation lives in [`apps/web/src/lib/atmosphere.ts`](https://github.com/Vatsal-Jha256/eldermere/blob/main/apps/web/src/lib/atmosphere.ts).

## Visual Generation

The room background is layered from several systems:

1. CSS gradient and glow layers from the selected palette.
2. A seeded canvas backdrop for terrain, structure, and grain.
3. Procedural mood overlays for weather, mist, caves, courts, and other biomes.
4. Short-lived visual event bursts for combat, recruitment, story steps, and errors.

The browser uses:

- `FastNoiseLite` for terrain and flowing motion
- seeded pseudo-random generators for repeatability
- simple cellular automata for cave and void-like spaces
- structured pattern overlays for court-like rooms

The effect is that a room can feel distinct from its metadata alone.

## Audio Generation

Ambient audio is generated from the same profile.

- rain rooms add softer noise and rain-like pulses
- wind rooms add low moving layers
- fire rooms add crackle and ember layers
- water rooms emphasize fluid motion
- sacred rooms add softer tonal layers
- void rooms lean darker and more sparse
- court rooms lean measured and ceremonial

Event cues also use the current room profile and command type to trigger short musical or noise-based responses.

The implementation lives in [`apps/web/src/lib/audio.ts`](https://github.com/Vatsal-Jha256/eldermere/blob/main/apps/web/src/lib/audio.ts).

## Why This Matters For Contributors

This system gives contributors a good payoff for a small amount of content work.

- Change the room `atmosphere` block and you change the visual and audio mood.
- Add a story arc with meaningful `room_hint` and `commands`, and the runtime can guide the player.
- Add a new pack interaction and another myth pack can react to it without extra code.

That means content authors can contribute useful work without touching the engine for every new idea.

## Probability Generation

The gameplay probability layer is deterministic math around random rolls, not hidden balancing magic.

- `fight` and `recruit` roll d20 checks.
- The server computes exact success chances by enumerating the possible d20 outcomes.
- Advantage and disadvantage are handled by enumerating the 400 two-die outcomes.
- Natural 1 and natural 20 rules are included in both the roll result and the displayed odds.
- Story variation tags are chosen uniformly when an arc starts.

Use `odds` while tuning rooms. If a DC feels wrong in play, the command gives a concrete success percentage before you change the content pack.

## Practical Guidelines

- Prefer a clear `palette`, `weather`, and `myth_layer` over empty atmosphere metadata.
- Use motifs that are concrete and readable by both the browser and the player.
- Keep room ids stable once a room is public.
- Match the atmosphere to the room function. Courts, vaults, rivers, and ferries should not all share the same mood profile.
- If you add a new visible event type, make sure it has a readable text cue and a sensible visual effect.
