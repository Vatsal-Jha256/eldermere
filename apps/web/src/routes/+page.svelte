<script lang="ts">
  import { onMount } from 'svelte';

  type RoomView = {
    id: string;
    name: string;
    description: string;
    exits: Record<string, string>;
  };

  type ServerEvent = {
    type: string;
    text: string;
    room?: RoomView;
  };

  const commands = ['quest', 'look', 'go east', 'recruit', 'go down', 'take', 'inventory', 'go west'];

  let command = $state('');
  let connected = $state(false);
  let room = $state<RoomView | null>(null);
  let log = $state([
    'Opening a path to the Eldermere server...'
  ]);
  let socket: WebSocket | null = null;
  const apiBase = import.meta.env.PUBLIC_API_BASE ?? 'http://localhost:8080';

  onMount(() => {
    const playerID = getPlayerID();
    socket = new WebSocket(toWebSocketURL(apiBase, '/ws', playerID));

    socket.addEventListener('open', () => {
      connected = true;
    });

    socket.addEventListener('message', (event) => {
      const parsed = parseServerEvent(event.data);
      if (parsed.room) {
        room = parsed.room;
      }
      log = [...log, parsed.text];
    });

    socket.addEventListener('close', () => {
      connected = false;
      log = [...log, 'Disconnected from the server.'];
    });

    socket.addEventListener('error', () => {
      log = [...log, 'Connection error. Is the Go server running on port 8080?'];
    });

    return () => {
      socket?.close();
    };
  });

  function submitCommand() {
    const trimmed = command.trim();
    if (!trimmed) return;

    log = [...log, `> ${trimmed}`];
    socket?.send(JSON.stringify({ command: trimmed }));
    command = '';
  }

  function parseServerEvent(data: string): ServerEvent {
    try {
      return JSON.parse(data) as ServerEvent;
    } catch {
      return { type: 'system', text: data };
    }
  }

  function getPlayerID() {
    const key = 'eldermere.player_id';
    const existing = localStorage.getItem(key);
    if (existing) return existing;

    const next = crypto.randomUUID();
    localStorage.setItem(key, next);
    return next;
  }

  function toWebSocketURL(base: string, path: string, playerID: string) {
    const url = new URL(path, base);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.searchParams.set('player_id', playerID);
    return url.toString();
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
      <p class="eyebrow">{room?.name ?? 'Connecting'} / Camelot Underbelly</p>
      <h1>Eldermere</h1>
      <p class="lede">
        {room?.description ??
          "A browser MUD for connected legends. Start in Arthur's Britain, recruit strange allies, and follow rumors that should not know each other yet."}
      </p>
    </div>
  </section>

  <section class="console" aria-label="Command console">
    <div class:online={connected} class="status">
      {connected ? 'Connected' : 'Disconnected'}
    </div>

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
        disabled={!connected}
      />
      <button type="submit" disabled={!connected}>Send</button>
    </form>

    <div class="chips" aria-label="Example commands">
      {#each commands as item}
        <button type="button" onclick={() => (command = item)}>{item}</button>
      {/each}
    </div>
  </section>
</main>
