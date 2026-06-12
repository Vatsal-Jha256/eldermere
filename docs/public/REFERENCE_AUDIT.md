# Reference Audit

Eldermere keeps local reference clones outside git in `reference-repos/`.

## Local Reference Repos

- `reference-repos/talesmud`
  - Upstream: https://github.com/TalesMUD/talesmud
  - License observed locally: MIT
  - Useful references: browser-first MUD flow, Go backend layering, WebSocket game loop, service/repository separation, content interfaces.

- `reference-repos/evennia`
  - Upstream: https://github.com/evennia/evennia
  - License observed locally: BSD 3-Clause
  - Useful references: game-agnostic MUD architecture, command/object/channel concepts, builder workflows, documentation depth, server/runtime separation.

## Reuse Rules

- Prefer original Eldermere code and data models unless a specific upstream implementation is intentionally reused.
- If code is reused from an upstream project, preserve the required license notice and document the source in the commit or a dedicated attribution file.
- Architecture, terminology, and workflow ideas can be adapted freely, but do not copy protected project branding, prose, assets, or game content.
- Use TalesMUD and Evennia to improve Eldermere's MUD architecture, not to become a fork of either project.

## Wider Inspiration

TalesMUD's own acknowledgments point toward useful ecosystem references:

- MUD history: decades of text-based virtual worlds and command-driven multiplayer spaces.
- `gopher-lua`: Lua scripting inside Go, worth revisiting if Eldermere adds mod scripting.
- `xterm.js`: terminal emulator for the web, worth evaluating if Eldermere's browser console needs richer terminal behavior.
- Svelte: already aligned with Eldermere's web client direction.
- Gin: useful Go HTTP reference, though Eldermere currently uses the standard library router.

These are not all Eldermere dependencies today. They are reference points to evaluate deliberately when a feature needs them.

## Current Direction

Near-term engine ideas to adapt:

- Clear command routing with explicit player, room, and broadcast audiences.
- Builder-friendly content validation before runtime loading.
- Room, object, story, faction, and channel concepts that stay data-first.
- Service boundaries that keep WebSockets, persistence, and game rules testable.
- Browser-first play while leaving room for traditional MUD concepts later.
