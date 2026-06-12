<script lang="ts">
  import { onMount } from 'svelte';

  type RoomView = {
    id: string;
    name: string;
    description: string;
    exits: Record<string, string>;
    atmosphere: {
      palette?: string;
      weather?: string;
      myth_layer?: string;
      motifs?: string[];
    };
  };

  type ServerEvent = {
    type: string;
    text: string;
    room?: RoomView;
  };

  type PlayerSession = {
    player_id: string;
    display_name: string;
    token: string;
  };

  const commands = [
    'help',
    'help story',
    'quest',
    'story',
    'story start sword-test',
    'story status',
    'travel arthurian-core',
    'map',
    'fight',
    'factions'
  ];

  let command = $state('');
  let connected = $state(false);
  let room = $state<RoomView | null>(null);
  let backgroundCanvas = $state<HTMLCanvasElement | null>(null);
  let log = $state([
    'Opening a path to the Eldermere server...'
  ]);
  let socket: WebSocket | null = null;
  const apiBase = import.meta.env.PUBLIC_API_BASE ?? 'http://localhost:8080';
  const atmosphereStyle = $derived(buildAtmosphereStyle(room));

  onMount(() => {
    let active = true;
    connect().catch((error) => {
      if (active) {
        log = [...log, `Connection setup failed: ${error instanceof Error ? error.message : 'unknown error'}`];
      }
    });

    return () => {
      active = false;
      socket?.close();
    };
  });

  $effect(() => {
    const canvas = backgroundCanvas;
    const currentRoom = room;
    if (!canvas) return;

    let frame = 0;
    const draw = () => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => drawAtmosphereCanvas(canvas, currentRoom));
    };

    draw();
    const observer = new ResizeObserver(draw);
    observer.observe(canvas);

    return () => {
      cancelAnimationFrame(frame);
      observer.disconnect();
    };
  });

  async function connect() {
    const session = await getPlayerSession();
    socket = new WebSocket(toWebSocketURL(apiBase, '/ws', session));

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
  }

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

  async function getPlayerSession() {
    const key = 'eldermere.session';
    const existing = localStorage.getItem(key);
    if (existing) {
      return JSON.parse(existing) as PlayerSession;
    }

    const response = await fetch(new URL('/api/v1/sessions', apiBase), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ display_name: 'Wanderer' })
    });

    if (!response.ok) {
      throw new Error(`session request failed with ${response.status}`);
    }

    const session = (await response.json()) as PlayerSession;
    localStorage.setItem(key, JSON.stringify(session));
    return session;
  }

  function toWebSocketURL(base: string, path: string, session: PlayerSession) {
    const url = new URL(path, base);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.searchParams.set('player_id', session.player_id);
    url.searchParams.set('token', session.token);
    return url.toString();
  }

  function buildAtmosphereStyle(current: RoomView | null) {
    const palette = paletteFor(current?.atmosphere?.palette);
    const motifSeed = hashText(current?.atmosphere?.motifs?.join('|') ?? current?.id ?? 'eldermere');
    const mistAngle = 25 + (motifSeed % 80);
    const glowX = 18 + (motifSeed % 64);
    const glowY = 12 + ((motifSeed >> 3) % 48);

    return [
      `--bg-a: ${palette[0]}`,
      `--bg-b: ${palette[1]}`,
      `--bg-c: ${palette[2]}`,
      `--mist-angle: ${mistAngle}deg`,
      `--glow-x: ${glowX}%`,
      `--glow-y: ${glowY}%`
    ].join(';');
  }

  function paletteFor(name?: string) {
    const palettes: Record<string, [string, string, string]> = {
      'rain-gold': ['#0e1612', '#7f6a32', '#d9b45f'],
      blackwater: ['#090d10', '#13232a', '#6b7f87'],
      'candle-smoke': ['#17110d', '#6f3e24', '#d9a45d'],
      'tavern-red': ['#190c0c', '#612018', '#d67b45'],
      'avalon-green': ['#081411', '#174736', '#96c7a1'],
      'relic-vault': ['#100f15', '#594c7a', '#c4a45d'],
      'coin-shadow': ['#11100b', '#5c4a1f', '#c59a3a'],
      'oracle-blue': ['#07131f', '#164466', '#8fc6d9'],
      'bronze-ash': ['#15100d', '#694421', '#c28f52']
    };
    return palettes[name ?? ''] ?? ['#101511', '#314439', '#e2b65f'];
  }

  function hashText(value: string) {
    let hash = 0;
    for (let index = 0; index < value.length; index += 1) {
      hash = (hash * 31 + value.charCodeAt(index)) >>> 0;
    }
    return hash;
  }

  function drawAtmosphereCanvas(canvas: HTMLCanvasElement, current: RoomView | null) {
    const rect = canvas.getBoundingClientRect();
    const pixelRatio = window.devicePixelRatio || 1;
    const width = Math.max(1, Math.floor(rect.width * pixelRatio));
    const height = Math.max(1, Math.floor(rect.height * pixelRatio));
    if (canvas.width !== width || canvas.height !== height) {
      canvas.width = width;
      canvas.height = height;
    }

    const context = canvas.getContext('2d');
    if (!context) return;

    context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    const w = rect.width;
    const h = rect.height;
    const palette = paletteFor(current?.atmosphere?.palette);
    const colors = palette.map(hexToRgb);
    const seed = hashText([
      current?.id ?? 'eldermere',
      current?.atmosphere?.weather ?? '',
      current?.atmosphere?.myth_layer ?? '',
      current?.atmosphere?.motifs?.join('|') ?? ''
    ].join('|'));
    const random = mulberry32(seed);

    const sky = context.createLinearGradient(0, 0, w, h);
    sky.addColorStop(0, rgb(colors[0]));
    sky.addColorStop(0.62, rgb(mix(colors[0], colors[1], 0.62)));
    sky.addColorStop(1, rgb(colors[1]));
    context.fillStyle = sky;
    context.fillRect(0, 0, w, h);

    drawGlow(context, w, h, colors[2], 0.18 + random() * 0.14, 0.16 + random() * 0.18, Math.min(w, h) * 0.5);
    drawGlow(context, w, h, colors[1], 0.72 + random() * 0.16, 0.2 + random() * 0.18, Math.min(w, h) * 0.38);
    drawHorizon(context, w, h, colors, random);
    drawMotifs(context, w, h, colors, current?.atmosphere?.motifs ?? [], random);
    drawWeather(context, w, h, current?.atmosphere?.weather ?? '', colors, random);
    drawGrain(context, w, h, colors[2], random);

    const vignette = context.createRadialGradient(w * 0.45, h * 0.42, Math.min(w, h) * 0.1, w * 0.5, h * 0.5, Math.max(w, h) * 0.75);
    vignette.addColorStop(0, 'rgba(0, 0, 0, 0)');
    vignette.addColorStop(1, 'rgba(0, 0, 0, 0.72)');
    context.fillStyle = vignette;
    context.fillRect(0, 0, w, h);
  }

  function drawGlow(context: CanvasRenderingContext2D, w: number, h: number, color: RGB, x: number, y: number, radius: number) {
    const glow = context.createRadialGradient(w * x, h * y, 0, w * x, h * y, radius);
    glow.addColorStop(0, rgba(color, 0.48));
    glow.addColorStop(0.45, rgba(color, 0.14));
    glow.addColorStop(1, rgba(color, 0));
    context.fillStyle = glow;
    context.fillRect(0, 0, w, h);
  }

  function drawHorizon(context: CanvasRenderingContext2D, w: number, h: number, colors: RGB[], random: () => number) {
    context.save();
    context.fillStyle = rgba(mix(colors[0], colors[1], 0.42), 0.72);
    context.beginPath();
    context.moveTo(0, h * 0.74);
    for (let x = 0; x <= w; x += Math.max(18, w / 18)) {
      const y = h * (0.58 + random() * 0.14);
      context.lineTo(x, y);
    }
    context.lineTo(w, h);
    context.lineTo(0, h);
    context.closePath();
    context.fill();

    context.globalAlpha = 0.28;
    for (let i = 0; i < 8; i += 1) {
      const y = h * (0.48 + i * 0.055);
      context.strokeStyle = rgba(colors[2], 0.2 - i * 0.015);
      context.lineWidth = 1;
      context.beginPath();
      context.moveTo(0, y + random() * 18);
      context.bezierCurveTo(w * 0.28, y - 22, w * 0.7, y + 24, w, y + random() * 18);
      context.stroke();
    }
    context.restore();
  }

  function drawMotifs(context: CanvasRenderingContext2D, w: number, h: number, colors: RGB[], motifs: string[], random: () => number) {
    const text = motifs.join(' ').toLowerCase();
    context.save();
    context.fillStyle = rgba(colors[0], 0.62);
    context.strokeStyle = rgba(colors[2], 0.34);
    context.lineWidth = 2;

    if (text.includes('banner') || text.includes('claim') || text.includes('table')) {
      for (let i = 0; i < 4; i += 1) drawBanner(context, w * (0.14 + i * 0.13), h * (0.38 + random() * 0.18), h * 0.26, colors[2]);
    }
    if (text.includes('candle') || text.includes('wax')) {
      for (let i = 0; i < 10; i += 1) drawCandle(context, w, h, w * (0.12 + random() * 0.72), h * (0.58 + random() * 0.24), 18 + random() * 42, colors[2]);
    }
    if (text.includes('bridge') || text.includes('jetty') || text.includes('ferry')) {
      drawBridge(context, w, h, colors[0], colors[2]);
    }
    if (text.includes('boat') || text.includes('shore') || text.includes('avalon')) {
      for (let i = 0; i < 3; i += 1) drawBoat(context, w * (0.28 + random() * 0.52), h * (0.58 + random() * 0.18), w * (0.08 + random() * 0.05), colors[0], colors[2]);
    }
    if (text.includes('stone') || text.includes('vault') || text.includes('mount')) {
      drawStone(context, w * 0.66, h * 0.58, Math.min(w, h) * 0.16, colors[0], colors[2]);
    }

    context.restore();
  }

  function drawBanner(context: CanvasRenderingContext2D, x: number, y: number, height: number, accent: RGB) {
    context.strokeStyle = rgba(accent, 0.35);
    context.beginPath();
    context.moveTo(x, y - height * 0.45);
    context.lineTo(x, y + height * 0.48);
    context.stroke();
    context.fillStyle = rgba(accent, 0.16);
    context.beginPath();
    context.moveTo(x, y - height * 0.42);
    context.lineTo(x + height * 0.28, y - height * 0.35);
    context.lineTo(x + height * 0.2, y - height * 0.13);
    context.lineTo(x, y - height * 0.18);
    context.closePath();
    context.fill();
  }

  function drawCandle(context: CanvasRenderingContext2D, w: number, h: number, x: number, y: number, height: number, accent: RGB) {
    context.fillStyle = 'rgba(15, 10, 6, 0.62)';
    context.fillRect(x - 3, y - height, 6, height);
    drawGlow(context, w, h, accent, x / w, (y - height) / h, height * 2.8);
  }

  function drawBridge(context: CanvasRenderingContext2D, w: number, h: number, dark: RGB, accent: RGB) {
    context.strokeStyle = rgba(accent, 0.24);
    context.lineWidth = Math.max(3, w * 0.006);
    context.beginPath();
    context.moveTo(w * 0.06, h * 0.64);
    context.quadraticCurveTo(w * 0.5, h * 0.5, w * 0.94, h * 0.66);
    context.stroke();
    context.fillStyle = rgba(dark, 0.54);
    context.fillRect(0, h * 0.67, w, h * 0.06);
  }

  function drawBoat(context: CanvasRenderingContext2D, x: number, y: number, width: number, dark: RGB, accent: RGB) {
    context.fillStyle = rgba(dark, 0.62);
    context.strokeStyle = rgba(accent, 0.3);
    context.beginPath();
    context.moveTo(x - width * 0.5, y);
    context.quadraticCurveTo(x, y + width * 0.28, x + width * 0.5, y);
    context.quadraticCurveTo(x, y + width * 0.12, x - width * 0.5, y);
    context.fill();
    context.stroke();
  }

  function drawStone(context: CanvasRenderingContext2D, x: number, y: number, radius: number, dark: RGB, accent: RGB) {
    context.fillStyle = rgba(dark, 0.5);
    context.strokeStyle = rgba(accent, 0.26);
    context.lineWidth = 2;
    context.beginPath();
    context.ellipse(x, y, radius * 0.7, radius, -0.08, 0, Math.PI * 2);
    context.fill();
    context.stroke();
  }

  function drawWeather(context: CanvasRenderingContext2D, w: number, h: number, weather: string, colors: RGB[], random: () => number) {
    const lower = weather.toLowerCase();
    context.save();
    if (lower.includes('rain') || lower.includes('drizzle')) {
      context.strokeStyle = rgba(colors[2], 0.18);
      context.lineWidth = 1;
      for (let i = 0; i < 90; i += 1) {
        const x = random() * w;
        const y = random() * h;
        context.beginPath();
        context.moveTo(x, y);
        context.lineTo(x - 18, y + 44);
        context.stroke();
      }
    } else {
      context.strokeStyle = rgba(colors[2], 0.1);
      for (let i = 0; i < 28; i += 1) {
        const y = random() * h;
        context.beginPath();
        context.moveTo(0, y);
        context.bezierCurveTo(w * 0.26, y - 26, w * 0.6, y + 34, w, y + random() * 20);
        context.stroke();
      }
    }
    context.restore();
  }

  function drawGrain(context: CanvasRenderingContext2D, w: number, h: number, accent: RGB, random: () => number) {
    const count = Math.min(900, Math.floor((w * h) / 1800));
    context.fillStyle = rgba(accent, 0.08);
    for (let i = 0; i < count; i += 1) {
      context.fillRect(random() * w, random() * h, 1, 1);
    }
  }

  type RGB = [number, number, number];

  function hexToRgb(hex: string): RGB {
    const normalized = hex.replace('#', '');
    return [
      Number.parseInt(normalized.slice(0, 2), 16),
      Number.parseInt(normalized.slice(2, 4), 16),
      Number.parseInt(normalized.slice(4, 6), 16)
    ];
  }

  function mix(a: RGB, b: RGB, amount: number): RGB {
    return [
      Math.round(a[0] + (b[0] - a[0]) * amount),
      Math.round(a[1] + (b[1] - a[1]) * amount),
      Math.round(a[2] + (b[2] - a[2]) * amount)
    ];
  }

  function rgb(color: RGB) {
    return `rgb(${color[0]}, ${color[1]}, ${color[2]})`;
  }

  function rgba(color: RGB, alpha: number) {
    return `rgba(${color[0]}, ${color[1]}, ${color[2]}, ${alpha})`;
  }

  function mulberry32(seed: number) {
    return function random() {
      let value = (seed += 0x6D2B79F5);
      value = Math.imul(value ^ (value >>> 15), value | 1);
      value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
      return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
    };
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
  <section class="room" aria-label="Current room" style={atmosphereStyle}>
    <canvas class="room__canvas" bind:this={backgroundCanvas} aria-hidden="true"></canvas>
    <div class="room__background"></div>
    <div class="room__content">
      <p class="eyebrow">
        {room?.name ?? 'Connecting'} / {room?.atmosphere?.myth_layer ?? 'Camelot Underbelly'}
      </p>
      <h1>Eldermere</h1>
      <p class="lede">
        {room?.description ??
          "A browser MUD for connected legends. Start in Arthur's Britain, recruit strange allies, and follow rumors that should not know each other yet."}
      </p>
      {#if room?.atmosphere?.weather || room?.atmosphere?.motifs?.length}
        <p class="atmosphere">
          {room.atmosphere.weather}
          {#if room.atmosphere.weather && room.atmosphere.motifs?.length} / {/if}
          {room.atmosphere.motifs?.join(', ')}
        </p>
      {/if}
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
