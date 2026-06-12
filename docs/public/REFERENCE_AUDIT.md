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

- `reference-repos/dikumud`
  - Upstream: https://github.com/Seifert69/DikuMUD
  - License observed locally: root LGPL 2.1 text plus `doc/license.doc` noting the 2020 LGPL release and preserving the original license for reference.
  - Useful references: historical room files, command conventions, combat loop, credits/help traditions.

- `reference-repos/circlemud`
  - Upstream mirror: https://github.com/Yuffster/CircleMUD
  - License observed locally: older CircleMUD/Diku license text with noncommercial and credit requirements.
  - Use as architecture/history reference only unless a specific compatible source release is selected and notices are handled.

- `reference-repos/fluffos`
  - Upstream: https://github.com/fluffos/fluffos
  - License observed locally: root `Copyright` carries LPmud/MudOS noncommercial restrictions; bundled third-party components have separate licenses.
  - Use as architecture/history reference for driver/mudlib separation, LPC runtime ideas, WebSocket support, and test organization.

- `reference-repos/ldmud`
  - Upstream: https://github.com/ldmud/ldmud
  - License observed locally: `COPYRIGHT` allows use/modification/redistribution but includes a no-monetary-gain restriction.
  - Use as architecture/history reference for driver design, documentation, and mudlib support.

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

Historical MUD references to study carefully:

- DikuMUD: historically important room/combat/content style. Treat as design history unless using a clearly compatible current source release with notices. References: https://dikumud.com/dikumud-license/ and `reference-repos/dikumud`.
- CircleMUD: useful for compact command-loop and builder workflow ideas; the checked mirror has older noncommercial requirements, so use as architecture/history reference by default. References: https://www.circlemud.org/license.html and `reference-repos/circlemud`.
- LPMud and modern LPMud drivers such as LDMud/FluffOS: useful for driver/mudlib separation and scripting architecture. The checked LDMud and FluffOS roots retain no-monetary-gain restrictions, so use as architecture/history reference by default. References: https://www.ldmud.eu/, https://github.com/fluffos/fluffos, `reference-repos/ldmud`, and `reference-repos/fluffos`.
- Merc, ROM, SMAUG, and Diku-family descendants: useful for command conventions, help systems, and area-file traditions, but reuse must respect derivative license chains and attribution requirements. Reference: https://github.com/alexmchale/merc-mud
- DGD: useful reference for evented driver design and LPC-family architecture. Reference: https://www.dworkin.nl/dgd/

Do not import code from historical MUD codebases until the exact source, license, and required notices are recorded.

## Current Direction

Near-term engine ideas to adapt:

- Clear command routing with explicit player, room, and broadcast audiences.
- Builder-friendly content validation before runtime loading.
- Room, object, story, faction, and channel concepts that stay data-first.
- Service boundaries that keep WebSockets, persistence, and game rules testable.
- Browser-first play while leaving room for traditional MUD concepts later.
