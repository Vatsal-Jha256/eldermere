<script lang="ts">
  const commands = ['look', 'go north', 'talk merlin', 'recruit', 'quest'];

  let command = $state('');
  let log = $state([
    'The rain over Camelot tastes faintly of iron.',
    'A cracked milestone points toward Avalon, Londinium, and somewhere the map refuses to name.',
    'Type a command to begin.'
  ]);

  function submitCommand() {
    const trimmed = command.trim();
    if (!trimmed) return;

    log = [...log, `> ${trimmed}`, commandResponse(trimmed)];
    command = '';
  }

  function commandResponse(input: string) {
    const lower = input.toLowerCase();
    if (lower === 'look') {
      return 'You stand in the Lantern Yard, where squires trade rumors beside a shrine of old river-stone.';
    }
    if (lower.startsWith('go')) {
      return 'The route is marked, but the server command loop will decide where it leads next.';
    }
    if (lower.startsWith('talk')) {
      return 'A hooded advisor smiles like he already knows which dice you will roll.';
    }
    if (lower === 'recruit') {
      return 'A half-wild oath spirit watches you. It may join you if the odds bend kindly.';
    }
    if (lower === 'quest') {
      return 'Quest seed: recover a stolen Excalibur fragment before it is sold beneath the old bridge.';
    }
    return 'The world listens. This command will become real once the backend parser lands.';
  }
</script>

<svelte:head>
  <title>Eldermere</title>
  <meta
    name="description"
    content="A browser MUD creature-RPG for Arthurian legend and connected myths."
  />
</svelte:head>

<main class="shell">
  <section class="room" aria-label="Current room">
    <div class="room__background"></div>
    <div class="room__content">
      <p class="eyebrow">Lantern Yard / Camelot Underbelly</p>
      <h1>Eldermere</h1>
      <p class="lede">
        A browser MUD for connected legends. Start in Arthur's Britain, recruit strange allies,
        and follow rumors that should not know each other yet.
      </p>
    </div>
  </section>

  <section class="console" aria-label="Command console">
    <div class="console__log" aria-live="polite">
      {#each log as line}
        <p>{line}</p>
      {/each}
    </div>

    <form class="command" onsubmit={(event) => { event.preventDefault(); submitCommand(); }}>
      <label for="command">Command</label>
      <input
        id="command"
        bind:value={command}
        autocomplete="off"
        spellcheck="false"
        placeholder="try: look"
      />
      <button type="submit">Send</button>
    </form>

    <div class="chips" aria-label="Example commands">
      {#each commands as item}
        <button type="button" onclick={() => (command = item)}>{item}</button>
      {/each}
    </div>
  </section>
</main>

